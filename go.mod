module github.com/aretext/aretext

go 1.15

require (
	github.com/gdamore/tcell v1.4.0
	github.com/mattn/go-runewidth v0.0.10
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.7.0
	golang.org/x/text v0.3.5
)

replace github.com/gdamore/tcell => github.com/wedaly/tcell v1.1.2-0.20200814032409-69f20cafeda4

replace github.com/mattn/go-runewidth => github.com/wedaly/go-runewidth v0.0.10-0.20200814015138-9d7065f24c9d
