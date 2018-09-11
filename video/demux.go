package video

import (
	"io"
	"log"
	"github.com/tigerjang/go-libav/avutil"
	"github.com/tigerjang/go-libav/avformat"
	"github.com/tigerjang/go-libav/avcodec"
	"syscall"
	"../io"
	"../utils"
)

func OpenAvFmtCtx(input dshRepIO.Accessor) (*avformat.Context, error) {
	decFmt, err := avformat.NewContextForInput()
	if err != nil {
		return nil, utils.LibavError(err, "Failed to create input avformat context!")
	}

	srcURI := ""
	bufferSize := 5 * 1024

	if input.Type() == dshRepIO.AccessFileSystem {
		srcURI = input.URI().Path()
	} else {
		// ******************* Custom IO Context *******************
		inIOCtx, err := avformat.NewCustomIOContext(
			bufferSize, 0,
			func(buffer []byte, size int) int {
				//fmt.Printf("### Buffer size: %d\n", len(buffer))
				nRead, err := input.Read(buffer)
				if err != nil { //遇到任何错误立即返回，并忽略 EOF 错误信息
					//fmt.Printf("### - Read Error: %v\n", err)
					if err == io.EOF {
						return avformat.AVErrorEOF
					}
					return -1
				}
				//fmt.Printf("### - Read %d bytes\n", nRead)
				input.Comment() // TODO
				return nRead
			},
			nil,
			func(offset int64, whence int) int64 {
				//fmt.Printf("@@@ Seek to: %d", offset)
				//fmt.Printf(", Whence: %d\n", whence)
				if whence == avformat.SeekWhenceSize {
					nowPos, _ := input.Seek(0, io.SeekCurrent)
					//fmt.Printf("@@@ - Now Position: %d", nowPos)
					fileSize, _ := input.Seek(0, io.SeekEnd)
					input.Seek(nowPos, io.SeekStart)
					//fmt.Printf(", Seek file size: %d\n", fileSize)
					return fileSize
				} else {
					pos, err := input.Seek(offset, whence)
					if err != nil { //遇到任何错误立即返回，并忽略 EOF 错误信息
						if err == io.EOF {
							return int64(avformat.AVErrorEOF)
						}
						return -1
					}
					input.Comment() // TODO
					//fmt.Printf("@@@ - Return Pos: %d\n", pos)
					return pos
				}
			})
		if err != nil {
			defer decFmt.Free()
			return nil, utils.LibavError(err, "Failed to create custom IO context!")
		}
		decFmt.SetIOContext(inIOCtx)
		decFmt.SetFlags(decFmt.Flags() | avformat.ContextFlagCustomIO)
		// ******************* Custom IO Context *******************
	}

	// set some options for opening file TODO
	//options := avutil.NewDictionary()
	//defer options.Free()
	//if err := options.Set("scan_all_pmts", "1"); err != nil {
	//	log.Fatalf("Failed to set input options: %v\n", err)
	//}

	if err = decFmt.OpenInput(srcURI, nil, nil); err != nil {
		defer decFmt.Free()
		return nil, utils.LibavError(err, "Failed to create custom IO context!")
		log.Fatalf("Failed to open input file: %v\n", err)
	}

	return decFmt, nil
	//if err = decFmt.FindStreamInfo(nil); err != nil {
	//	log.Fatalf("Failed to find stream info: %v\n", err)
	//}
}

func trySetupHWDecode()  {
	hwTypeName := "dxva2"
	hwType := avutil.HWDeviceFindTypeByName(hwTypeName)
	if hwType == avutil.HWDeviceTypeNONE {
		log.Fatalf("Cannot support '%s' in this example.\n", hwTypeName)
	}
	//C.hw_pix_fmt = findFmtByHWType(hwType)  // TODO
	C.find_fmt_by_hw_type((C.enum_AVHWDeviceType)(hwType))
	syscall.Ne
}


type Demuxer interface {
	Seek (sec float32)
	GetPacket () *avcodec.Packet
	SeekFrame (frameNum int)
	// cache
}

type LocalDemuxer struct {
	avFmtCtx *avformat.Context
}

func NewLocalDemuxer(input interface{}) *LocalDemuxer {
	fmtCtx, err := OpenAvFmtCtx(input)

	return &LocalDemuxer{avFmtCtx:fmtCtx}
}



type RemoteDemuxer struct {

}

