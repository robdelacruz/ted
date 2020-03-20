all: ted

dep:
	go get -u github.com/nsf/termbox-go

ted: ted.go buf.go bufiterch.go bufiterwl.go buf_test.go editview.go label.go panel.go prompt.go tabs.go tbutil.go textsurface.go widget.go
	go build -o ted ted.go buf.go bufiterch.go bufiterwl.go buf_test.go editview.go label.go panel.go prompt.go tabs.go tbutil.go textsurface.go widget.go

clean:
	rm -rf ted

