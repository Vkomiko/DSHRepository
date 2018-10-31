package main

/*
#include "demuxer.h"
#cgo pkg-config: libavformat libavutil libavcodec
 */
import "C"

import (
	"errors"
	"fmt"
	"github.com/tigerjang/DSHRepository/utils"
	"log"
	"runtime"
	"syscall"
	"unsafe"
)

// https://ffmpeg.org/doxygen/4.0/hw_decode_8c-example.html
// https://godoc.org/github.com/golang/protobuf/proto

type MediaType int // C.enum_AVMediaType

const (
	MediaTypeUnknown    MediaType = C.AVMEDIA_TYPE_UNKNOWN
	MediaTypeVideo      MediaType = C.AVMEDIA_TYPE_VIDEO
	MediaTypeAudio      MediaType = C.AVMEDIA_TYPE_AUDIO
	MediaTypeData       MediaType = C.AVMEDIA_TYPE_DATA
	MediaTypeSubtitle   MediaType = C.AVMEDIA_TYPE_SUBTITLE
	MediaTypeAttachment MediaType = C.AVMEDIA_TYPE_ATTACHMENT
)

type CodecParameters struct {
	avCodecPara *C.AVCodecParameters
}  // TODO free

type Demuxer interface {
	NumStream () uint
	GetCodecParameters (streamId uint) *CodecParameters
	GetBestStream (streamType MediaType) int
	BestStreamsMap () map[MediaType]int
	ReadPacket (pkt *Packet) (*Packet, error)
	PickPacket (pkt *Packet, filter []bool) (*Packet, error)
}

func PickPacket (d Demuxer, pkt *Packet, filter []bool) (*Packet, error) {
	if rp, err := d.ReadPacket(pkt); err == nil {
		sid := rp.avPacket.stream_index
		if filter[sid] {
			return pkt, nil
		}
		pkt.UnRef()
		return nil, nil
	} else {
		return nil, err
	}
}

//----------------------------------------------------------------------------------------------------------------------

type FFmpegErrorType int

const (
	FFmpegError FFmpegErrorType = iota + 0x10000
)

type DSHFFmpegError struct {
	utils.DSHStdError
}

func (t FFmpegErrorType) New(c C.int, msg string) error {
	code := int(C.GO_AVERROR(c))
	return &DSHFFmpegError{utils.DSHStdError{code, int(t), msg}}
}

func (t FFmpegErrorType) FromErrno(eno syscall.Errno, msg string) error {
	code := C.GO_AVERROR(C.int(eno))
	return t.New(code, msg)
}

func (e *DSHFFmpegError) Error() string {
	ffMsg := "Unknown error"
	size := C.size_t(256)
	buf := (*C.char)(C.av_mallocz(size))
	defer C.av_free(unsafe.Pointer(buf))
	if C.av_strerror(C.int(e.Ecode), buf, size-1) == 0 {
		ffMsg = C.GoString(buf)
	}
	return fmt.Sprintf("FFmpegError [%d] %s\n    FFmpeg ErrorMsg: %s", e.Ecode, e.Emsg, ffMsg)
}

//----------------------------------------------------------------------------------------------------------------------
// https://ffmpeg.org/doxygen/4.0/group__lavc__packet.html
type Packet struct {
	avPacket *C.AVPacket
}

func (p *Packet) finalize () {
	C.av_packet_free(&p.avPacket)
}

func __init_Packet__ (p *Packet) *Packet {
	runtime.SetFinalizer(p, (*Packet).finalize)
	return p
}

func NewPacket() (*Packet, error) {
	if avp := C.av_packet_alloc(); unsafe.Pointer(avp) != C.NULL {
		return __init_Packet__(&Packet{avp}), nil
	}
	return nil, utils.InitializeError.New(1, "Packet allocation error !")
}

// Allocate the payload of a packet and $TODO initialize its fields with default values ???
func (p *Packet) AllocPayload (size int) error {
	if rc := C.av_new_packet(p.avPacket, C.int(size)); rc != 0 {
		return FFmpegError.New(rc, "Allocate Packet Payload Error.") // TODO  !!!!!!!!!!!!!!!!!
	}
	return nil
}

func (p *Packet) UnRef () {
	C.av_packet_unref(p.avPacket)
}

//----------------------------------------------------------------------------------------------------------------------

type FFmpegDemuxer struct {
	ffDemuxer *C.FFDemuxer
}

func NewFFmpegDemuxer(uri string) (*FFmpegDemuxer, error) {
	cFilename := C.CString(uri)
	defer C.free(unsafe.Pointer(cFilename))

	var ffDemuxer *C.FFDemuxer
	if ret := C.ff_new_demuxer(&ffDemuxer, cFilename, (*C.FFIOContext)(unsafe.Pointer(nil))); ret < 0 {
		return nil, errors.New("hhhhhh") // TODO
	}

	d := &FFmpegDemuxer{ffDemuxer}
	runtime.SetFinalizer(d, (*FFmpegDemuxer).finalize)
	return d, nil
}

func (d *FFmpegDemuxer) finalize() {
	C.ff_free_demuxer(d.ffDemuxer)
}

func (d *FFmpegDemuxer) NumStream () uint {
	return uint(d.ffDemuxer.avf_ctx.nb_streams)
}

func (d *FFmpegDemuxer) GetCodecParameters (streamId uint) *CodecParameters {
	var paras *C.AVCodecParameters
	if ret := C.ff_demuxer_get_codec_paras(d.ffDemuxer, &paras, C.uint(streamId)); ret != 0 {
		return nil
	}
	return &CodecParameters{paras}
}

func (d *FFmpegDemuxer) GetBestStream (streamType MediaType) int {
	if streamType < C.NUM_MEDIA_TYPE {
		return int(d.ffDemuxer.best_streams[C.int(streamType)])
	}
	return -1
}

func (d *FFmpegDemuxer) BestStreamsMap () map[MediaType]int {
	m := make(map[MediaType]int)
	for i := 0; i < C.NUM_MEDIA_TYPE; i++ {
		s := d.ffDemuxer.best_streams[i]
		if s >=0 {
			m[MediaType(i)] = int(s)
		}
	}
	return m
}

func (d *FFmpegDemuxer) ReadPacket (pkt *Packet) (*Packet, error) {
	if pkt == nil {
		var err error
		if pkt, err = NewPacket(); err != nil {
			return nil, err
		}
	} else {
		pkt.UnRef()  // FixMe  Here or outside ?
	}
	if rc := C.av_read_frame(d.ffDemuxer.avf_ctx, pkt.avPacket); rc < 0 {
		return nil, FFmpegError.New(rc, "Can not read packet.") // TODO: case C.AVERROR_EOF:
	}
	return pkt, nil
}

func (d *FFmpegDemuxer) PickPacket (pkt *Packet, filter []bool) (*Packet, error) {
	return PickPacket(d, pkt, filter)
}

//----------------------------------------------------------------------------------------------------------------------
// https://ffmpeg.org/doxygen/4.0/group__lavu__frame.html

type Frame struct {
	avFrame *C.AVFrame
}

func (f *Frame) finalize () {
	C.av_frame_free(&f.avFrame)
}

func __init_Frame__ (f *Frame) *Frame {
	runtime.SetFinalizer(f, (*Frame).finalize)
	return f
}

func NewFrame () (*Frame, error) {
	if avf := C.av_frame_alloc(); unsafe.Pointer(avf) != C.NULL {
		return __init_Frame__(&Frame{avf}), nil
	}
	return nil, utils.InitializeError.New(1, "Frame allocation error !")
}

func (f *Frame) UnRef () {
	C.av_frame_unref(f.avFrame)
}

//----------------------------------------------------------------------------------------------------------------------

type FFmpegDecoder struct {
	ffDecoder *C.FFDecoder
}

func NewFFmpegDecoder(codec_paras *CodecParameters) (*FFmpegDecoder, error) {
	var ffDecoder *C.FFDecoder
	if ret := C.ff_new_decoder(&ffDecoder, codec_paras.avCodecPara); ret < 0 {
		return nil, errors.New("hhhhhh") // TODO
	}

	d := &FFmpegDecoder{ffDecoder}
	runtime.SetFinalizer(d, (*FFmpegDecoder).finalize)
	return d, nil
}

func (d *FFmpegDecoder) finalize () {
	C.ff_free_decoder(d.ffDecoder)
}

func (d *FFmpegDecoder) Push (pkt *Packet) error {
	if rc := C.avcodec_send_packet(d.ffDecoder.avc_ctx, pkt.avPacket); rc < 0 {
		return FFmpegError.New(rc, "Fail to Push Packet into Decoder.")
	}
	return nil
}

func (d *FFmpegDecoder) Pop (frm *Frame) (*Frame, error) {
	if frm == nil {
		var err error
		if frm, err = NewFrame(); err != nil {
			return nil, err
		}
	} else {
		frm.UnRef()  // FixMe  Here or outside ?
	}
	switch rc := C.avcodec_receive_frame(d.ffDecoder.avc_ctx, frm.avFrame); C.GO_AVERROR(rc) {
	case C.AVERROR_EOF:
		return nil, utils.CMsgEOF
	case C.EAGAIN:
		return nil, utils.CMsgStarve
	case 0:
		return frm, nil
	default:
		return nil, FFmpegError.New(rc, "Receive Frame from Decoder Error.")
	}
}

func main()  {
	cDeviceType := C.CString("d3d11va")  // TODO
	defer C.free(unsafe.Pointer(cDeviceType))

	uri := ""

	demuxer, _ := NewFFmpegDemuxer(uri)
	fmt.Println(demuxer.NumStream())

	paras := demuxer.GetCodecParameters(0)
	fmt.Println(paras)

	decoder, _ := NewFFmpegDecoder(paras)
	fmt.Println(decoder)

	fmt.Println(demuxer.GetBestStream(MediaTypeAudio))
	fmt.Println(demuxer.BestStreamsMap())

	packet, err := NewPacket()
	if err != nil {
		fmt.Println(err)
	}

	fil := make([]bool, demuxer.NumStream())
	fil[demuxer.GetBestStream(MediaTypeVideo)] = true
	fmt.Println(fil)

	var frame *Frame
	if frame, err = NewFrame(); err != nil{
		fmt.Println(err)
	}

	var pkt *Packet
	var frm *Frame
	for i := 0; i < 50; i ++ {
		pkt = nil
		for pkt == nil {
			if pkt, err = demuxer.PickPacket(packet, fil); err != nil {
				log.Fatal(err)
			}
		}

		if err = decoder.Push(pkt); err != nil {
			log.Fatal(err)
		}
		frm = frame

		for frm != nil {
			if frm, err = decoder.Pop(frame); err != nil {
				if _, ok := err.(utils.ControlMsg); !ok {
					log.Fatal(err)
				}
			}
			a := frm
			fmt.Println(a, 2333)
		}
	}


}
