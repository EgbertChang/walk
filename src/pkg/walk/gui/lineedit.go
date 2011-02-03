// Copyright 2010 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gui

import (
	"os"
	"syscall"
	"unsafe"
)

import (
	"walk/drawing"
	. "walk/winapi"
	. "walk/winapi/user32"
)

var lineEditSubclassWndProcPtr uintptr
var lineEditOrigWndProcPtr uintptr

func lineEditSubclassWndProc(hwnd HWND, msg uint, wParam, lParam uintptr) uintptr {
	le, ok := widgetsByHWnd[hwnd].(*LineEdit)
	if !ok {
		return CallWindowProc(lineEditOrigWndProcPtr, hwnd, msg, wParam, lParam)
	}

	return le.wndProc(hwnd, msg, wParam, lParam, lineEditOrigWndProcPtr)
}

type LineEdit struct {
	Widget
}

func newLineEdit(parentHWND HWND) (*LineEdit, os.Error) {
	if lineEditSubclassWndProcPtr == 0 {
		lineEditSubclassWndProcPtr = syscall.NewCallback(lineEditSubclassWndProc)
	}

	hWnd := CreateWindowEx(
		WS_EX_CLIENTEDGE, syscall.StringToUTF16Ptr("EDIT"), nil,
		ES_AUTOHSCROLL|WS_CHILD|WS_TABSTOP|WS_VISIBLE,
		0, 0, 120, 24, parentHWND, 0, 0, nil)
	if hWnd == 0 {
		return nil, lastError("CreateWindowEx")
	}

	le := &LineEdit{Widget: Widget{hWnd: hWnd}}

	var succeeded bool
	defer func() {
		if !succeeded {
			le.Dispose()
		}
	}()

	lineEditOrigWndProcPtr = uintptr(SetWindowLong(hWnd, GWL_WNDPROC, int(lineEditSubclassWndProcPtr)))
	if lineEditOrigWndProcPtr == 0 {
		return nil, lastError("SetWindowLong")
	}

	le.SetFont(defaultFont)

	widgetsByHWnd[hWnd] = le

	succeeded = true

	return le, nil
}

func NewLineEdit(parent IContainer) (*LineEdit, os.Error) {
	if parent == nil {
		return nil, newError("parent cannot be nil")
	}

	le, err := newLineEdit(parent.Handle())
	if err != nil {
		return nil, err
	}

	var succeeded bool
	defer func() {
		if !succeeded {
			le.Dispose()
		}
	}()

	le.parent = parent
	if err = parent.Children().Add(le); err != nil {
		return nil, err
	}

	succeeded = true

	return le, nil
}

func (le *LineEdit) CueBanner() (string, os.Error) {
	buf := make([]uint16, 128)
	if FALSE == SendMessage(le.hWnd, EM_GETCUEBANNER, uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf))) {
		return "", newError("EM_GETCUEBANNER failed")
	}

	return syscall.UTF16ToString(buf), nil
}

func (le *LineEdit) SetCueBanner(value string) os.Error {
	if FALSE == SendMessage(le.hWnd, EM_SETCUEBANNER, FALSE, uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(value)))) {
		return newError("EM_SETCUEBANNER failed")
	}

	return nil
}

func (*LineEdit) LayoutFlags() LayoutFlags {
	return ShrinkHorz | GrowHorz
}

func (le *LineEdit) PreferredSize() drawing.Size {
	return le.dialogBaseUnitsToPixels(drawing.Size{50, 14})
}

func (le *LineEdit) wndProc(hwnd HWND, msg uint, wParam, lParam uintptr, origWndProcPtr uintptr) uintptr {
	switch msg {
	case WM_GETDLGCODE:
		if wParam == VK_RETURN {
			return DLGC_WANTALLKEYS
		}
	}

	return le.Widget.wndProc(hwnd, msg, wParam, lParam, lineEditOrigWndProcPtr)
}
