package gst



/*
#cgo pkg-config: gstreamer-1.0 gstreamer-base-1.0 gstreamer-app-1.0 gstreamer-plugins-base-1.0 gstreamer-video-1.0 gstreamer-audio-1.0 gstreamer-plugins-bad-1.0
#include "gst.h"
*/
import "C"


import  (
	"fmt"
	"unsafe"
	"errors"
	"reflect"
	"runtime"
)



type PadAddedCallback func(name string, element *Element, pad *Pad)



type StateOptions int

const (
	StateVoidPending StateOptions = C.GST_STATE_VOID_PENDING
	StateNull        StateOptions = C.GST_STATE_NULL
	StateReady       StateOptions = C.GST_STATE_READY
	StatePaused      StateOptions = C.GST_STATE_PAUSED
	StatePlaying     StateOptions = C.GST_STATE_PLAYING
)


type Element struct {

	GstElement *C.GstElement
	onPadAdded PadAddedCallback
}


func (e *Element) Name() (name string) {
	n := (*C.char)(unsafe.Pointer(C.gst_object_get_name((*C.GstObject)(unsafe.Pointer(e.GstElement)))))
	if n != nil {
		name = string(nonCopyCString(n, C.int(C.strlen(n))))
	}

	return
}


func (e *Element) Link(dst *Element) bool {

	result := C.gst_element_link(e.GstElement, dst.GstElement)
	if result == C.TRUE {
		return true
	}
	return false
}


func (e *Element) GetPadTemplate(name string) (padTemplate *PadTemplate) {

	n := (*C.gchar)(unsafe.Pointer(C.CString(name)))
	defer C.g_free(C.gpointer(unsafe.Pointer(n)))
	CPadTemplate := C.gst_element_class_get_pad_template(C.X_GST_ELEMENT_GET_CLASS(e.GstElement), n)
	padTemplate = &PadTemplate{
		C: CPadTemplate,
	}

	return
}

func (e *Element) GetRequestPad(padTemplate *PadTemplate, name string, caps *Caps) (pad *Pad) {

	var n *C.gchar
	var c *C.GstCaps

	if name == "" {
		n = nil
	} else {
		n = (*C.gchar)(unsafe.Pointer(C.CString(name)))
		defer C.g_free(C.gpointer(unsafe.Pointer(n)))
	}
	if caps == nil {
		c = nil
	} else {
		c = caps.caps
	}
	CPad := C.gst_element_request_pad(e.GstElement, padTemplate.C, n, c)
	pad = &Pad{
		pad: CPad,
	}

	return
}


func (e *Element) GetStaticPad(name string) (pad *Pad) {

	n := (*C.gchar)(unsafe.Pointer(C.CString(name)))
	defer C.g_free(C.gpointer(unsafe.Pointer(n)))
	CPad := C.gst_element_get_static_pad(e.GstElement, n)
	pad = &Pad{
		pad: CPad,
	}

	return
}


func (e *Element) AddPad(pad *Pad) bool {

	Cret := C.gst_element_add_pad(e.GstElement, pad.pad)
	if Cret == 1 {
		return true
	}

	return false
}



func (e *Element) GetClock() (gstClock *Clock) {

	CElementClock := C.gst_element_get_clock(e.GstElement)

	gstClock = &Clock{
		C: CElementClock,
	}

	runtime.SetFinalizer(gstClock, func(gstClock *Clock) {
		C.gst_object_unref(C.gpointer(unsafe.Pointer(gstClock.C)))
	})

	return
}


// appsrc 
func (e *Element) PushBuffer(buffer *Buffer) (err error) {

	// TODO
	// GST_IS_APP_SRC check
	var gstReturn C.GstFlowReturn

	gstReturn = C.gst_app_src_push_buffer((*C.GstAppSrc)(unsafe.Pointer(e.GstElement)), buffer.C)
	if buffer.C == nil {

	}
	if gstReturn != C.GST_FLOW_OK {
		err = errors.New("could not push buffer on appsrc element")
		return
	}

	return
}


// appsrc 
func (e *Element) PushSample(sample *Sample) (err error) {

	// TODO
	// GST_IS_APP_SRC check

	var gstReturn C.GstFlowReturn

	gstReturn = C.gst_app_src_push_sample((*C.GstAppSrc)(unsafe.Pointer(e.GstElement)), sample.C)
	if sample.C == nil {

	}
	if gstReturn != C.GST_FLOW_OK {
		err = errors.New("could not push sample on appsrc element")
		return
	}

	return
}

// appsink
func (e *Element) PullSample() (sample *Sample, err error) {

	// TODO
	// GST_IS_APP_SRC check

	CGstSample := C.gst_app_sink_pull_sample((*C.GstAppSink)(unsafe.Pointer(e.GstElement)))
	if CGstSample == nil {
		err = errors.New("could not pull a sample from appsink")
		return
	}
	CGstSampleCopy := C.gst_sample_copy(CGstSample)

	C.gst_sample_unref(CGstSample)

	var width, height C.gint
	CCaps := C.gst_sample_get_caps(CGstSampleCopy)
	CCStruct := C.gst_caps_get_structure(CCaps, 0)
	C.gst_structure_get_int(CCStruct, (*C.gchar)(unsafe.Pointer(C.CString("width"))), &width)
	C.gst_structure_get_int(CCStruct, (*C.gchar)(unsafe.Pointer(C.CString("height"))), &height)

	sample = &Sample{
		C:      CGstSampleCopy,
		Width:  uint32(width),
		Height: uint32(height),
	}

	runtime.SetFinalizer(sample, func(gstSample *Sample) {
		C.gst_sample_unref(gstSample.C)
	})

	return
}

//export go_callback_new_pad_thunk
func go_callback_new_pad_thunk(Cname *C.gchar, CgstElement *C.GstElement, CgstPad *C.GstPad, Cdata C.gpointer) {
	element := (*Element)(unsafe.Pointer(Cdata))
	callback := element.onPadAdded
	name := C.GoString((*C.char)(unsafe.Pointer(Cname)))
	// element := &GstElement{
	// 	gstElement: CgstElement,
	// }
	pad := &Pad{
		pad: CgstPad,
	}

	callback(name, element, pad)
}


func (e *Element) SetPadAddedCallback(callback PadAddedCallback)  {
	e.onPadAdded = callback

	detailedSignal := (*C.gchar)(unsafe.Pointer(C.CString("pad-added")))
	defer C.g_free(C.gpointer(unsafe.Pointer(detailedSignal)))
	C.X_g_signal_connect(e.GstElement, detailedSignal, (*[0]byte)(C.cb_new_pad), (C.gpointer)(unsafe.Pointer(e)))
}



func ElementFactoryMake(factoryName string, name string) (e *Element, err error) {
	var pName *C.gchar

	pFactoryName := (*C.gchar)(unsafe.Pointer(C.CString(factoryName)))
	defer C.g_free(C.gpointer(unsafe.Pointer(pFactoryName)))
	if name == "" {
		pName = nil
	} else {
		pName = (*C.gchar)(unsafe.Pointer(C.CString(name)))
		defer C.g_free(C.gpointer(unsafe.Pointer(pName)))
	}
	gstElt := C.gst_element_factory_make(pFactoryName, pName)

	if gstElt == nil {
		err = errors.New(fmt.Sprintf("could not create a GStreamer element factoryName %s, name %s", factoryName, name))
		return
	}

	e = &Element{
		GstElement: gstElt,
	}

	return
}


func nonCopyGoBytes(ptr uintptr, length int) []byte {
	var slice []byte
	header := (*reflect.SliceHeader)(unsafe.Pointer(&slice))
	header.Cap = length
	header.Len = length
	header.Data = ptr
	return slice
}

func nonCopyCString(data *C.char, size C.int) []byte {
	return nonCopyGoBytes(uintptr(unsafe.Pointer(data)), int(size))
}
