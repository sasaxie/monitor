package models

import "time"

var Program *ProgramInfo

func init() {
	Program = new(ProgramInfo)
	Program.Runtime = time.Now().UTC()
}

type ProgramInfo struct {
	Runtime time.Time
}
