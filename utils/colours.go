package utils

import (
	"github.com/fatih/color"
)

var (
	CIndex    = color.New(color.FgHiCyan)
	CSymbol   = color.New(color.FgHiMagenta)
	CTitle    = color.New(color.FgHiGreen).Add(color.Bold)
	CReadTime = color.New(color.FgHiMagenta)
	CURL      = color.New(color.FgHiYellow)
	CError    = color.New(color.FgHiRed)
	CExcerpt  = color.New(color.FgHiWhite)
	CTag      = color.New(color.FgHiBlue)

	CIndexSprint    = CIndex.SprintFunc()
	CSymbolSprint   = CSymbol.SprintFunc()
	CTitleSprint    = CTitle.SprintFunc()
	CReadTimeSprint = CReadTime.SprintFunc()
	CURLSprint      = CURL.SprintFunc()
	CErrorSprint    = CError.SprintFunc()
	CExcerptSprint  = CExcerpt.SprintFunc()
	CTagSprint      = CTag.SprintFunc()
)
