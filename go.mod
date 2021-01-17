module github.com/aretext/aretext

go 1.15

require (
	github.com/gdamore/tcell v1.4.0
	github.com/mattn/go-runewidth v0.0.10
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.7.0
	golang.org/x/text v0.3.5
)

replace (
	github.com/gdamore/tcell => github.com/aretext/tcell v1.4.1-0.20210117062323-2d397edb2c29
	github.com/mattn/go-runewidth => github.com/aretext/go-runewidth v0.0.11-0.20210117061314-7dc49ce56729
)
