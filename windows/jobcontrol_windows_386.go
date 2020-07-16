// ****************************************************************************
//
// This file is part of a MASA library or program.
// Refer to the included end-user license agreement for restrictions.
//
// Copyright (c) 2014 MASA Group
//
// ****************************************************************************
package windows

type JobObjectExtendedLimitInformation struct {
	BasicLimitInformation JobObjectBasicLimitInformation
	align1                uint32
	IoInfo                IoCounters
	ProcessMemoryLimit    uintptr // SIZE_T
	JobMemoryLimit        uintptr // SIZE_T
	PeakProcessMemoryUsed uintptr // SIZE_T
	PeakJobMemoryUsed     uintptr // SIZE_T
}
