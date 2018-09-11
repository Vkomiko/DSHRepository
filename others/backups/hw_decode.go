package main

/*
#include <stdio.h>
#include <libavcodec/avcodec.h>
#include <libavutil/pixfmt.h>
#include <libavutil/hwcontext.h>
// Fixme: temp !!!!!!!!!!!!!!!!
#include <libswscale/swscale.h>

static int gp_sws_scale_frame(struct SwsContext * ctx, AVFrame * src, AVFrame * dst) {
	return sws_scale(ctx,
		(const uint8_t* const*)src->data, src->linesize, 0, src->height,
		dst->data, dst->linesize);
	// TODO return
}

#cgo pkg-config: libavcodec libswscale libavutil
*/
import "C"

import (
	//"flag"
	"github.com/tigerjang/go-libav/avutil"
	"github.com/tigerjang/go-libav/avformat"
	"github.com/tigerjang/go-libav/avcodec"
	"github.com/tigerjang/go-libav/swscale"
	//"github.com/tigerjang/go-libav/avfilter"
)

import (
	"log"
	//"unsafe"
	"fmt"
	"os"
	//"io"
	"io"
)

func main() {
	decFmt, err := avformat.NewContextForInput()
	if err != nil {
		log.Fatalf("Failed to open input context: %v\n", err)
	}
	defer decFmt.Free()

	// ******************* Custom IO Context *******************
	fp := "./ttt.mkv"
	file, err := os.Open(fp)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(0)
	}
	defer file.Close()

	inIOCtx, err := avformat.NewCustomIOContext(
		1024 * 128, 0,
		func(buffer []byte, size int) int {
			//fmt.Printf("### Buffer size: %d\n", len(buffer))
			nRead, err := file.Read(buffer)
			if err != nil { //遇到任何错误立即返回，并忽略 EOF 错误信息
				//fmt.Printf("### - Read Error: %v\n", err)
				if err == io.EOF {
					return C.AVERROR_EOF
				}
				return -1
			}
			fmt.Printf("### - Read %d bytes\n", nRead)
			return nRead
		},
		nil,
		func(offset int64, whence int) int64 {
			fmt.Printf("@@@ Seek to: %d", offset)
			//fmt.Printf(", Whence: %d\n", whence)
			if whence == avformat.SeekWhenceSize {
				nowPos, _ := file.Seek(0, io.SeekCurrent)
				//fmt.Printf("@@@ - Now Position: %d", nowPos)
				fileSize, _ := file.Seek(0, io.SeekEnd)
				file.Seek(nowPos, io.SeekStart)
				//fmt.Printf(", Seek file size: %d\n", fileSize)
				return fileSize
			} else {
				pos, err := file.Seek(offset, whence)
				if err != nil { //遇到任何错误立即返回，并忽略 EOF 错误信息
					if err == io.EOF {
						return C.AVERROR_EOF
					}
					return -1
				}
				//fmt.Printf("@@@ - Return Pos: %d\n", pos)
				return pos
			}
		})
	if err != nil {
		log.Fatalf("Cannot Create IOContext '%s'\n", err)
	}
	decFmt.SetIOContext(inIOCtx)
	decFmt.SetFlags(decFmt.Flags() | avformat.ContextFlagCustomIO)
	// ******************* Custom IO Context *******************

	/* open the input file */
	//inputFileName := "./ttt.mkv"
	if err = decFmt.OpenInput("", nil, nil); err != nil {
		log.Fatalf("Failed to open input file: %v\n", err)
	}

	if err = decFmt.FindStreamInfo(nil); err != nil {
		log.Fatalf("Failed to find stream info: %v\n", err)
	}

	/* find the video stream information */
	streamId, codec, err := decFmt.FindBestStream(avutil.MediaTypeVideo, -1, -1, 0)
	if err != nil {
		log.Fatalf("Cannot find a video stream in the input file\n")
	}

	decCodec, err := avcodec.NewContextWithCodec(codec);
	if err != nil {
		log.Fatalf("Failed to create codec context: %v\n", err)
	}
	defer decCodec.Free()

	decStream := decFmt.Streams()[streamId]

	if err := decCodec.SetParameters(decStream.CodecParameters()); err != nil {
		log.Fatalf("Failed to copy codec context: %v\n", err)
	}

	if err := decCodec.SetInt64Option("refcounted_frames", 1); err != nil {
		log.Fatalf("Failed to copy codec context: %v\n", err)
	}

	// ******************* Setup HWDeviceContext *******************
	//hwTypeName := "dxva2"
	hwTypeName := "d3d11va"
	hwType := avcodec.HWDeviceFindTypeByName(hwTypeName)
	if hwType == avcodec.HWDeviceTypeNONE {
		log.Fatalf("Cannot support '%s' in this example.\n", hwTypeName)
	}

	hwDeviceCtx, err1 := avcodec.NewHWDeviceContext(hwType, "", nil, 0)
	if err1 != nil {
		log.Fatalf("Failed to init hardware device context")
	}
	defer hwDeviceCtx.Free()
	decCodec.SetHWDeviceContext(hwDeviceCtx)

	hwDevice2PxlFmtMap := map[string][]string {
		"vdpau": {"vdpau_h264", "vdpau_mpeg1", "vdpau_mpeg2", "vdpau_wmv3", "vdpau_vc1", "vdpau_mpeg4", "vdpau"},
		"cuda": {"cuda"},
		"vaapi": {"vaapi_moco", "vaapi_idct", "vaapi_vld", "vaapi"},
		"dxva2": {"dxva2_vld"},
		"qsv": {"qsv"},
		"videotoolbox": {"videotoolbox"},
		"d3d11va": {"d3d11", "d3d11va_vld"},
		"drm": {"drm_prime"},
		"none": {"none"},
	}

	hwPxlFmtTryList := hwDevice2PxlFmtMap[hwTypeName]

	decCodec.GetFormatCallback(func(cCtx *avcodec.Context, availPxlFmts []string) string {
		for _, pf := range availPxlFmts {
			for _, pf_ := range hwPxlFmtTryList {
				if pf == pf_ {
					return pf
				}
			}
		}
		log.Printf("Can not find proper pixel format for device type: %s", hwTypeName)
		log.Printf("Pixel format available: %s", availPxlFmts)
		return "none"  // TODO
	})
	// ******************* Setup HWDeviceContext *******************

	if err := decCodec.OpenWithCodec(codec, nil); err != nil {
		log.Fatalf("Failed to open codec: %v\n", err)
	}

	decPkt, err := avcodec.NewPacket()
	if err != nil {
		log.Fatalf("Failed to alloc decode packet: %v\n", err)
	}
	defer decPkt.Free()

	decFrame, err := avutil.NewFrame()
	if err != nil {
		log.Fatalf("Failed to alloc decode frame: %v\n", err)
	}
	defer decFrame.Free()

	swFrame, err := avutil.NewFrame()
	if err != nil {
		log.Fatalf("Failed to alloc decode frame: %v\n", err)
	}

	outFrame, err := avutil.NewFrame()
	if err != nil {
		log.Fatalf("Failed to alloc decode frame: %v\n", err)
	}
	//pf_bgra, _ := avutil.FindPixelFormatByName("bgra")
	pf_yuvj420p, _ := avutil.FindPixelFormatByName("yuvj420p")
	//swFrame.SetPixelFormat(pf_bgra)
	//log.Println(swFrame.PixelFormat().Name())

	var swsCtx *swscale.SwsContext
	defer swsCtx.Free()  // TODO  nil !!!!!!!!!!!!!!!!!!!

	doFirstFrame := func (fr *avutil.Frame) {
		outFrame.SetPixelFormat(pf_yuvj420p)
		//outFrame.SetPixelFormat(pf_bgra)
		outFrame.SetHeight(fr.Height())
		outFrame.SetWidth(fr.Width())
		outFrame.GetBuffer()
		swsCtx, err = swscale.GetContext(fr.Width(), fr.Height(), fr.PixelFormat(),
			fr.Width(), fr.Height(), outFrame.PixelFormat(),
			swscale.SWS_AREA, nil, nil, nil)
		if err != nil {
			log.Fatal("Get swscale context error !")
		}
	}
	isFirstFrame := true


	defer swFrame.Free()
	//////////////////////////////////////////////
	iframe := 1
	//var seekTS int64 = 2333 * 50
	var seekTS int64 = 2333 * 500

	if err := decFmt.SeekToTimestamp(decStream.Index(), seekTS - 666, seekTS, seekTS + 666, 0); err != nil {
		log.Printf("Failed to seek: %v\n", err)
	}

	decodeFrame := func () (int, error) {
		var frame *avutil.Frame
		if level, err := decCodec.ReceiveFrame(decFrame); level > 0 {
			return level, err
		}
		defer decFrame.Unref()
		if decFrame.PixelFormat() == decCodec.GetHwCtxPixelFormat() {
			/* retrieve data from GPU to CPU */
			fmts, _ := avcodec.HWFrameTransferGetFormats(
				decFrame.GetHWFramesCtx(), avcodec.HWFrameTransferDirectionFrom, 0)
			log.Println(fmts)
			for _f := range fmts {
				log.Println(avutil.PixelFormat(_f).Name())
			}
			if err := avcodec.HWFrameTransferData(swFrame, decFrame, 0); err != nil {
				return 2, err
			}
			frame = swFrame
			log.Println(decFrame.PixelFormat().Name())
			log.Println(swFrame.PixelFormat().Name())
			defer swFrame.Unref()
		} else {
			frame = decFrame
		}

		if isFirstFrame {
			doFirstFrame(frame)
			isFirstFrame = false
		}

		fmt.Println(frame.LineSize())
		fmt.Println(frame.Data())
		fmt.Println(outFrame.LineSize())
		swsCtx.SwsScale(frame.Data(), frame.LineSize(), 0, frame.Height(),
			outFrame.Data(), outFrame.LineSize())
		fmt.Println(outFrame.Data())
		fmt.Println(outFrame.LineSize())

		h := frame.Height()
		h += 0
		saveFrame(outFrame, 0, 0, iframe, decFrame)

		iframe ++

		//log.Printf("--- Current TimeStamp: %d\n", decPkt.PTS())  // TODO /////////////////////
		//if err := decFmt.SeekToTimestamp(decStream.Index(), seekTS - 666, seekTS, seekTS + 666, 0); err != nil {
		//	log.Printf("Failed to seek: %v\n", err)
		//}
		seekTS += 2333 * 50
		//fmt.Printf("\r@@@ iFrame: %d", iframe)
		return 0, nil
	}

	decodePacket := func () (stop bool, retErr error) {
		_, err := decFmt.ReadFrame(decPkt)
		if err != nil {
			log.Fatalf("Failed to read packet: %v\n", err)
			return true, err
		}
		defer decPkt.Unref()

		if decPkt.StreamIndex() != decStream.Index() {
			return false, nil
		}

		decPkt.Print()
		if (int(decPkt.Flags()) & 1) == 1 {
			seekTS += 0
		}

		err = decCodec.SendPacket(decPkt)
		if err != nil {
			log.Printf("@@@ Failed to send packet: %v\n", err)
			return true, err
		}
		for true {
			if level, err := decodeFrame(); level > 0 {
				if level > 1 {
					log.Printf("@@@ Failed to receive packet: %v\n", err)
					if level > 2 {
						return true, err
					}
				}
				break
			}
		}
		return false, nil
	}

	for true {
		if stop, _ := decodePacket(); stop {
			break
		}

	}
}

func saveFrame(frame *avutil.Frame, width, height, idx int, decFrame *avutil.Frame) error {
	fp := fmt.Sprintf("./out/%d_%d.jpeg", decFrame.PacketPTS(), int(decFrame.PictureType()))

	outFmt := avformat.GuessOutputFromFileName(fp)
	if outFmt == nil {
		log.Fatalf("Failed to guess output for output file: %s\n", fp)
	}

	outFmtCtx, err := avformat.NewContextForOutput(outFmt);
	if err != nil {
		log.Fatalf("Failed to open output context: %v\n", err)
	}
	outFmtCtx.SetFileName(fp)
	defer outFmtCtx.Free()

	// prepare first video stream for encoding
	encStream, err := outFmtCtx.NewStreamWithCodec(nil)
	if err != nil {
		log.Fatalf("Failed to open output video stream: %v\n", err)
	}

	codecID := outFmtCtx.Output().VideoCodecID()
	//codecID := outFmt.GuessCodecID(fp, codecCtx.CodecType())
	encoder := avcodec.FindEncoderByID(codecID)
	//log.Printf("Encoder name: %s", encoder.Name())  // TODO ///////////////////////////////////////////
	if encoder == nil {
		log.Fatalf("Could not find encoder: %v\n", codecID)
	}

	codecCtx := encStream.CodecContext()
	//codecCtx, err := avcodec.NewContextWithCodec(encoder)
	//if  err != nil {
	//	log.Fatalf("Failed to create codec context: %v\n", err)
	//}
	//defer codecCtx.Free() // TODO: need ???

	//encodeCtx, err := avcodec.NewContextWithCodec(encoder)
	//if  err != nil {
	//	log.Fatalf("Failed to create codec context: %v\n", err)
	//}

	codecCtx.SetCodecType(avutil.MediaTypeVideo)
	codecCtx.SetCodecID(codecID)
	codecCtx.SetWidth(frame.Width())
	codecCtx.SetHeight(frame.Height())
	//codecCtx.SetWidth(1024)
	//codecCtx.SetHeight(1024)
	//codecCtx.SetPixelFormat((avutil.PixelFormat)(C.AV_PIX_FMT_YUVJ420P))  // 对mjpeg encoder不能乱设
	codecCtx.SetPixelFormat(frame.PixelFormat())
	timebase := avutil.NewRational(1, 25) // TODO free
	codecCtx.SetTimeBase(timebase)
	//codecCtx.SetFrameRate(avutil.NewRational(25, 1))
	//codecCtx.SetBitRate(400000)
	//codecCtx.SetGOPSize(10)
	//codecCtx.SetMaxBFrames(1)

	if err = codecCtx.OpenWithCodec(encoder, nil); err != nil {
		log.Fatalf("Failed to open codec: %v\n", err)
	}

	//outFmtCtx.Dump(0, fp, false)

	// prepare I/O
	flags := avformat.IOFlagWrite
	outIO, err := avformat.OpenIOContext(fp, flags, nil, nil);
	if err != nil {
		log.Fatalf("Failed to open I/O context: %v\n", err)
	}
	outFmtCtx.SetIOContext(outIO)
	defer outIO.Close()

	///////////////
	if err := outFmtCtx.WriteHeader(nil); err != nil {
		log.Fatalf("Failed to write header: %v\n", err)
	}

	encPkt, err := avcodec.NewPacket()
	if err != nil {
		log.Fatalf("Failed to alloc decode packet: %v\n", err)
	}
	defer encPkt.Free()
	//encPkt.SetSize(xxx)

	if err := codecCtx.SendFrame(frame); err != nil {
		//log.Fatalf("Failed to send frame: %v\n", err)  // TODO ////////////////////////////////////////
	}

	for true {
		if err := codecCtx.ReceivePacket(encPkt); err != nil {
			//log.Printf("Failed to send frame: %v\n", err)  // TODO ////////////////////////////////////////
			// TODO 判断是结束还是错误!!!!!!!!!
			// 结束: ret == AVERROR(EAGAIN) || ret == AVERROR_EOF
			break
		}

		encPkt.SetPosition(-1)
		encPkt.SetStreamIndex(encStream.Index())
		//encPkt.RescaleTime(ctx.encCodec.TimeBase(), ctx.encStream.TimeBase())
		if err := outFmtCtx.InterleavedWriteFrame(encPkt); err != nil {
			log.Fatalf("Failed to write packet: %v\n", err)
		}

		//if err := outFmtCtx.WriteFrame(encPkt); err != nil {
		//	log.Fatalf("Failed to write packet: %v\n", err)
		//}  TODO: 区别???

		encPkt.Unref()
	}

	if err := outFmtCtx.WriteTrailer(); err != nil {
		log.Fatalf("Failed to write header: %v\n", err)
	}

	sss := outIO.Size()
	sss += 0

	return nil
}




