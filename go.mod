module github.com/wedaly/aretext

go 1.14

require (
	github.com/gdamore/tcell v1.3.1-0.20200620214305-79a04f102133
	github.com/mattn/go-runewidth v0.0.9
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.6.1
)

replace github.com/gdamore/tcell => github.com/wedaly/tcell v1.1.2-0.20200814032409-69f20cafeda4

replace github.com/mattn/go-runewidth => github.com/wedaly/go-runewidth v0.0.10-0.20200814015138-9d7065f24c9d
