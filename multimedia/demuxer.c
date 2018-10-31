//
// Created by tigerjang on 2018/10/30.
//

#include "demuxer.h"

static AVBufferRef *hw_device_ctx = NULL;  // TODO
static enum AVPixelFormat hw_pix_fmt;  // TODO
//static FILE *output_file = NULL;

static int hw_decoder_init(AVCodecContext *ctx, const enum AVHWDeviceType type)
{
    int err = 0;
    if ((err = av_hwdevice_ctx_create(&hw_device_ctx, type,
                                      NULL, NULL, 0)) < 0) {
        fprintf(stderr, "Failed to create specified HW device.\n");
        return err;
    }
    ctx->hw_device_ctx = av_buffer_ref(hw_device_ctx);
    return err;
}

static enum AVPixelFormat get_hw_format(AVCodecContext *ctx,
                                        const enum AVPixelFormat *pix_fmts)
{
    const enum AVPixelFormat *p;
    for (p = pix_fmts; *p != -1; p++) {
        if (*p == hw_pix_fmt)
            return *p;
    }
    fprintf(stderr, "Failed to get HW surface format.\n");
    return AV_PIX_FMT_NONE;
}

int ff_new_demuxer (FFDemuxer ** ret, const char * uri, FFIOContext * io_ctx) {
    FFDemuxer * demuxer;
    AVFormatContext * avf_ctx = NULL;
    int i;

    /* open the input file */
    if (avformat_open_input(&avf_ctx, uri, NULL, NULL) != 0) {
        fprintf(stderr, "Cannot open input uri '%s'\n", uri);
        return -1;
    }

    /* find the video stream information */
    if (avformat_find_stream_info(avf_ctx, NULL) < 0) {
        fprintf(stderr, "Cannot find input stream information.\n");
        return -1;
    }

    demuxer = (FFDemuxer *)malloc(sizeof(FFDemuxer));

    /* find the best streams ? */
    for (i = 0; i < NUM_MEDIA_TYPE; i++) {
        if (i == AVMEDIA_TYPE_VIDEO || i == AVMEDIA_TYPE_AUDIO) {  // TODO: More?
            demuxer -> best_streams[i] = av_find_best_stream(avf_ctx, i, -1, -1, NULL, 0);
        }
        else {
            demuxer -> best_streams[i] = -1;
        }
    }

    demuxer -> avf_ctx = avf_ctx;
    *ret = demuxer;
    return 0;
}

void ff_free_demuxer (FFDemuxer * d) {
    avformat_close_input(&(d -> avf_ctx));
    free(d);
}

int ff_demuxer_get_codec_paras (FFDemuxer * d, AVCodecParameters ** ret, unsigned int stream_id) {
    if (stream_id < d -> avf_ctx -> nb_streams) {
        *ret = d -> avf_ctx -> streams[stream_id] -> codecpar;
        return 0;
    }
    return -1;
}

//----------------------------------------------------------------------------------------------------------------------

int ff_new_decoder (FFDecoder ** ret, AVCodecParameters * codec_paras) {
    FFDecoder * decoder;
    AVCodec * codec = NULL;  // TODO: Free ???
    AVCodecContext * avc_ctx = NULL;
    int rc;

    // AVCodecParameters --avcodec_find_decoder()--> AVCodec
    codec = avcodec_find_decoder(codec_paras -> codec_id);
    if (codec == NULL) {
        return -1;
    }

    // type_name ---av_hwdevice_find_type_by_name()--> enum_AVHWDeviceType

    // AVCodec, enum_AVHWDeviceType --avcodec_get_hw_config()--> AVCodecHWConfig

    // AVCodecHWConfig.pix_fmt --> AVPixelFormat

    // AVCodec --avcodec_alloc_context3()--> AVCodecContext
    if (!(avc_ctx = avcodec_alloc_context3(codec)))
        return AVERROR(ENOMEM);

    // AVCodecParameters ==avcodec_parameters_to_context()==> AVCodecContext
    if (avcodec_parameters_to_context(avc_ctx, codec_paras) < 0)
        return -1;

    // SET func get_format ==> AVCodecContext  ==> AVPixelFormat

    //  av_opt_set() ............  TODO ext options
    av_opt_set_int(avc_ctx, "refcounted_frames", 1, 0);

    // enum_AVHWDeviceType ==av_hwdevice_ctx_create()==> AVCodecContext.hw_device_ctx

    // AVCodec ==avcodec_open2==> AVCodecContext
    if ((rc = avcodec_open2(avc_ctx, codec, NULL)) < 0) {
        fprintf(stderr, "Failed to open codec for stream #%u\n", 233);
        return -1;
    }

    decoder = (FFDecoder *)malloc(sizeof(FFDecoder));
    decoder -> avc_ctx = avc_ctx;
    *ret = decoder;
    return 0;
}

void ff_free_decoder (FFDecoder * d) {
    avcodec_free_context(&(d -> avc_ctx));
    free(d);
}

//
//int new_demuxer (const char * filename, const char * device_type) {
//    int video_stream, ret;
//    AVStream *video = NULL;
//
//
//    AVPacket packet;
//    enum AVHWDeviceType type;
//    int i;
//
//    type = av_hwdevice_find_type_by_name(device_type);
//    if (type == AV_HWDEVICE_TYPE_NONE) {
//        fprintf(stderr, "Device type %s is not supported.\n", device_type);
//        fprintf(stderr, "Available device types:");
//        while((type = av_hwdevice_iterate_types(type)) != AV_HWDEVICE_TYPE_NONE)
//            fprintf(stderr, " %s", av_hwdevice_get_type_name(type));
//        fprintf(stderr, "\n");
//        return -1;
//    }
//
//    \\\\\\\
//
//    /* find the video stream information */
//     video_stream = ret;
//
//    for (i = 0;; i++) {
//        const AVCodecHWConfig *config = avcodec_get_hw_config(decoder, i);
//        if (!config) {
//            fprintf(stderr, "Decoder %s does not support device type %s.\n",
//                    decoder->name, av_hwdevice_get_type_name(type));
//            return -1;
//        }
//        if (config->methods & AV_CODEC_HW_CONFIG_METHOD_HW_DEVICE_CTX &&
//            config->device_type == type) {
//            hw_pix_fmt = config->pix_fmt;
//            break;
//        }
//    }
//
//
//    video = input_ctx->streams[video_stream];
//
//    decoder_ctx->get_format  = get_hw_format;
//
//    if (hw_decoder_init(decoder_ctx, type) < 0)
//        return -1;
//
//
//    /* actual decoding and dump the raw data */
//    while (ret >= 0) {
//        if ((ret = av_read_frame(input_ctx, &packet)) < 0)
//            break;
//        if (video_stream == packet.stream_index) {
////            ret = decode_write(decoder_ctx, &packet);
//        }
//        av_packet_unref(&packet);  // TODO
//    }
//
//    /* flush the decoder */
////    packet.data = NULL;
////    packet.size = 0;
////    ret = decode_write(decoder_ctx, &packet);
////    av_packet_unref(&packet);
////    if (output_file)
////        fclose(output_file);
////
////    av_buffer_unref(&hw_device_ctx);
//
//    return 0;
//}

