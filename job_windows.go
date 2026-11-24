//go:build windows

package cove

import (
	"unsafe"

	"golang.org/x/sys/windows"
)

type limit struct {
	job windows.Handle
	ok  bool
}

func prepare(memory uint64) (*limit, error) {
	job, err := windows.CreateJobObject(nil, nil)
	if err != nil {
		if memory > 0 {
			return nil, ErrLimited
		}
		return &limit{}, nil
	}
	info := windows.JOBOBJECT_EXTENDED_LIMIT_INFORMATION{}
	info.BasicLimitInformation.LimitFlags = windows.JOB_OBJECT_LIMIT_KILL_ON_JOB_CLOSE
	if memory > 0 {
		info.BasicLimitInformation.LimitFlags |= windows.JOB_OBJECT_LIMIT_PROCESS_MEMORY | windows.JOB_OBJECT_LIMIT_JOB_MEMORY
		info.ProcessMemoryLimit = uintptr(memory)
		info.JobMemoryLimit = uintptr(memory)
	}
	ret, err := windows.SetInformationJobObject(
		job,
		windows.JobObjectExtendedLimitInformation,
		uintptr(unsafe.Pointer(&info)),
		uint32(unsafe.Sizeof(info)),
	)
	if ret == 0 || err != nil {
		_ = windows.CloseHandle(job)
		if memory > 0 {
			return nil, ErrLimited
		}
		return &limit{}, nil
	}
	return &limit{job: job, ok: true}, nil
}

func (l *limit) assign(pid int) error {
	if l == nil || !l.ok {
		return nil
	}
	access := uint32(windows.PROCESS_SET_QUOTA | windows.PROCESS_TERMINATE)
	handle, err := windows.OpenProcess(access, false, uint32(pid))
	if err != nil {
		return err
	}
	defer windows.CloseHandle(handle)
	return windows.AssignProcessToJobObject(l.job, handle)
}

func (l *limit) cancel() error {
	if l != nil && l.ok {
		return windows.TerminateJobObject(l.job, 1)
	}
	return nil
}

func (l *limit) close() {
	if l != nil && l.ok {
		_ = windows.CloseHandle(l.job)
	}
}
