package sample

import (
	"log"
	"os"
)

func badFunc() {
	panic("oops") // want `спользование panic запрещено`

	log.Fatal("fatal") // want `вызов log\.Fatal запрещён вне функции main пакета main`

	os.Exit(1) // want `вызов os\.Exit запрещён вне функции main пакета main`
}
