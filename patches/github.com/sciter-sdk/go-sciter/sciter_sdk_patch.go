package sciter

/*
#cgo CFLAGS: -Iinclude
#include "sciter-x.h"

extern BOOL SCAPI SciterProcX(HWINDOW hwnd, SCITER_X_MSG* pMsg);
*/
import "C"
import "unsafe"

func Wrap2(hwnd unsafe.Pointer) *Sciter {
	s := &Sciter{
		hwnd:      C.HWINDOW(hwnd),
		callbacks: make(map[*CallbackHandler]struct{}),
	}
	return s
}

func (e *Element) GetHandle() C.HELEMENT {
	return e.handle
}

type XMsg interface {
	GetCStruct() unsafe.Pointer
}

type XMsgPaint struct {
	xMsg *C.SCITER_X_MSG_PAINT
}

func (s *Sciter) XMsgPaint(layerElement *Element, foreLayer bool) *XMsgPaint {
	_foreLayer := C.FALSE
	if foreLayer {_foreLayer = C.TRUE}
	x := &XMsgPaint{
		&C.SCITER_X_MSG_PAINT{
			header: C.SCITER_X_MSG{msg: C.SXM_PAINT},
			element: layerElement.GetHandle(),
			isFore: C.INT(_foreLayer),
			targetType: C.SPT_DEFAULT,
		},
	}
	return x
}

func (s *XMsgPaint) GetCStruct() unsafe.Pointer {
	return unsafe.Pointer(s.xMsg)
}

func (s *Sciter) ProcX(pMsg XMsg) int {
	//C_pMsg := (*C.SCITER_X_MSG)(pMsg.GetCStruct())
	//println(C_pMsg.msg)
	return int(C.SciterProcX(s.hwnd, (*C.SCITER_X_MSG)(pMsg.GetCStruct())))
	//return int(C.SciterProcX(s.hwnd, &(*C.SCITER_X_MSG_PAINT)(pMsg.GetCStruct()).header))
	//return true
}

