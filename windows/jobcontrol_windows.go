// ****************************************************************************
//
// This file is part of a MASA library or program.
// Refer to the included end-user license agreement for restrictions.
//
// Copyright (c) 2014 MASA Group
//
// ****************************************************************************
package windows

import (
	"os"
	"syscall"
	"unsafe"
)

var (
	kernel32 = syscall.NewLazyDLL("kernel32.dll")

	procCreateJobObjectW         = kernel32.NewProc("CreateJobObjectW")
	procSetInformationJobObject  = kernel32.NewProc("SetInformationJobObject")
	procAssignProcessToJobObject = kernel32.NewProc("AssignProcessToJobObject")
)

func CreateJobObject(sa *syscall.SecurityAttributes, name *uint16) (syscall.Handle, error) {
	res, _, err := procCreateJobObjectW.Call(
		uintptr(unsafe.Pointer(sa)),
		uintptr(unsafe.Pointer(name)))
	if res == 0 {
		return syscall.InvalidHandle, os.NewSyscallError("CreateJobObject", err)
	}
	return syscall.Handle(res), nil
}

func SetInformationJobObject(job syscall.Handle, infoclass uint32,
	info unsafe.Pointer, length uint32) error {

	res, _, err := procSetInformationJobObject.Call(
		uintptr(job),
		uintptr(infoclass),
		uintptr(info),
		uintptr(length))

	if res == 0 {
		return os.NewSyscallError("SetInformationJobObject", err)
	}
	return nil
}

type JobObjectBasicAccountingInformation struct {
	TotalUserTime             uint64
	TotalKernelTime           uint64
	ThisPeriodTotalUserTime   uint64
	ThisPeriodTotalKernelTime uint64
	TotalPageFaultCount       uint32
	TotalProcesses            uint32
	ActiveProcesses           uint32
	TotalTerminatedProcesses  uint32
}

type JobObjectBasicUiRestrictions struct {
	UIRestrictionClass uint32
}

type JobObjectBasicLimitInformation struct {
	PerProcessUserTimeLimit uint64  // LARGE_INTEGER
	PerJobUserTimeLimit     uint64  // LARGE_INTEGER
	LimitFlags              uint32  // DWORD
	MinimumWorkingSetSize   uintptr // SIZE_T
	MaximumWorkingSetSize   uintptr // SIZE_T
	ActiveProcessLimit      uint32  // DWORD
	Affinity                uintptr // ULONG_PTR
	PriorityClass           uint32  // DWORD
	SchedulingClass         uint32  // DWORD
}

const (
	JOB_OBJECT_LIMIT_ACTIVE_PROCESS             = 8
	JOB_OBJECT_LIMIT_AFFINITY                   = 0x00000010
	JOB_OBJECT_LIMIT_DIE_ON_UNHANDLED_EXCEPTION = 0x400
	JOB_OBJECT_LIMIT_JOB_MEMORY                 = 0x200
	JOB_OBJECT_LIMIT_JOB_TIME                   = 4
	JOB_OBJECT_LIMIT_KILL_ON_JOB_CLOSE          = 0x2000
	JOB_OBJECT_LIMIT_PROCESS_MEMORY             = 0x100
	JOB_OBJECT_LIMIT_PROCESS_TIME               = 2
	JOB_OBJECT_LIMIT_WORKINGSET                 = 1
)

type IoCounters struct {
	ReadOperationCount  uint64 // ULONGLONG
	WriteOperationCount uint64 // ULONGLONG
	OtherOperationCount uint64 // ULONGLONG
	ReadTransferCount   uint64 // ULONGLONG
	WriteTransferCount  uint64 // ULONGLONG
	OtherTransferCount  uint64 // ULONGLONG
}

const (
	JobObjectExtendedLimitInformationClass = 9
)

func SetJobObjectExtendedLimitInformation(job syscall.Handle, info *JobObjectExtendedLimitInformation) error {
	return SetInformationJobObject(job, JobObjectExtendedLimitInformationClass,
		unsafe.Pointer(info), uint32(unsafe.Sizeof(*info)))
}

func AssignProcessToJobObject(job syscall.Handle, process syscall.Handle) error {
	res, _, err := procAssignProcessToJobObject.Call(
		uintptr(job),
		uintptr(process))
	if res == 0 {
		return os.NewSyscallError("AssignProcessToJobObject", err)
	}
	return nil
}

func CreateChildProcessKillingJob() (syscall.Handle, error) {
	job, err := CreateJobObject(nil, nil)
	if err != nil {
		return syscall.InvalidHandle, err
	}
	info := JobObjectExtendedLimitInformation{}
	info.BasicLimitInformation.LimitFlags = JOB_OBJECT_LIMIT_KILL_ON_JOB_CLOSE
	err = SetJobObjectExtendedLimitInformation(job, &info)
	if err != nil {
		_ = syscall.CloseHandle(job)
		return syscall.InvalidHandle, err
	}
	proc, err := syscall.GetCurrentProcess()
	if err != nil {
		_ = syscall.CloseHandle(job)
		return syscall.InvalidHandle, err
	}
	err = AssignProcessToJobObject(job, proc)
	if err != nil {
		_ = syscall.CloseHandle(job)
		return syscall.InvalidHandle, err
	}
	return job, nil
}

var (
	killJob syscall.Handle
)

func MakeProcessKillItsSubProcess() error {
	job, err := CreateChildProcessKillingJob()
	if err != nil {
		return err
	}
	killJob = job
	return nil
}
