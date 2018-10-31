//
// Created by tigerjang on 2018/10/30.
//

#include "demuxer.h"


static enum AVPixelFormat get_hw_format(AVCodecContext *ctx,
                                        const enum AVPixelFormat *pix_fmts)
{
    const enum AVPixelFormat *p;
    const enum AVPixelFormat tp = ((FFCodecCtxOpaque *)(ctx -> opaque)) -> hw_pix_fmt;
    for (p = pix_fmts; *p != -1; p++) {
        if (*p == tp)
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

int ff_new_decoder (FFDecoder ** ret, AVCodecParameters * codec_paras, const char * device_type) {
    FFDecoder * decoder;
    AVCodec * codec = NULL;  // TODO: Free ???
    AVCodecContext * avc_ctx = NULL;
    FFCodecCtxOpaque * ctx_opaque;
    AVBufferRef * hw_device_ctx = NULL;
    enum AVHWDeviceType dtype;
    int is_hw = 0;
    void * old_get_format;
    int rc, i;

    // AVCodecParameters --avcodec_find_decoder()--> AVCodec
    codec = avcodec_find_decoder(codec_paras -> codec_id);
    if (codec == NULL) {
        return -1;
    }

    // AVCodec --avcodec_alloc_context3()--> AVCodecContext
    if (!(avc_ctx = avcodec_alloc_context3(codec)))
        return AVERROR(ENOMEM);

    ctx_opaque = (FFCodecCtxOpaque *)malloc(sizeof(FFCodecCtxOpaque));
    avc_ctx -> opaque = (void *)ctx_opaque;

    // AVCodecParameters ==avcodec_parameters_to_context()==> AVCodecContext
    if (avcodec_parameters_to_context(avc_ctx, codec_paras) < 0)
        goto init_decoder_fail;

    //  av_opt_set() ............  TODO ext options
    av_opt_set_int(avc_ctx, "refcounted_frames", 1, 0);

    if (device_type != NULL) {
        // type_name ---av_hwdevice_find_type_by_name()--> enum_AVHWDeviceType
        dtype = av_hwdevice_find_type_by_name(device_type);
        if (dtype == AV_HWDEVICE_TYPE_NONE) {
            fprintf(stderr, "Device type %s is not supported.\n", device_type);
            fprintf(stderr, "Available device types:");
            while((dtype = av_hwdevice_iterate_types(dtype)) != AV_HWDEVICE_TYPE_NONE)
                fprintf(stderr, " %s", av_hwdevice_get_type_name(dtype));
            fprintf(stderr, "\n");
            goto hw_decode_fail;
        }

        // AVCodec, enum_AVHWDeviceType --avcodec_get_hw_config()--> AVCodecHWConfig
        for (i = 0;; i++) {
            const AVCodecHWConfig *config = avcodec_get_hw_config(codec, i);
            if (!config) {
                fprintf(stderr, "Decoder %s does not support device type %s.\n",
                        codec->name, av_hwdevice_get_type_name(dtype));
                goto hw_decode_fail;
            }
            if (config->methods & AV_CODEC_HW_CONFIG_METHOD_HW_DEVICE_CTX && config->device_type == dtype) {
                // AVCodecHWConfig.pix_fmt --> AVPixelFormat
                ctx_opaque -> hw_pix_fmt = config -> pix_fmt;
                break;
            }
        }

        // SET func get_format ==> AVCodecContext  ==> AVPixelFormat
        old_get_format = avc_ctx -> get_format;
        avc_ctx -> get_format = get_hw_format;

        // enum_AVHWDeviceType ==av_hwdevice_ctx_create()==> AVCodecContext.hw_device_ctx
        if ((rc = av_hwdevice_ctx_create(&hw_device_ctx, dtype, NULL, NULL, 0)) < 0) {
            fprintf(stderr, "Failed to create specified HW device.\n");
            goto hw_decode_fail;
        }
        avc_ctx -> hw_device_ctx = av_buffer_ref(hw_device_ctx);

        is_hw = 1;
        goto hw_decode_ok;

hw_decode_fail:
//        avc_ctx -> get_format = old_get_format;  // TODO Why Wrong ????
hw_decode_ok:
        i = 0;
    }

    // AVCodec ==avcodec_open2==> AVCodecContext
    if ((rc = avcodec_open2(avc_ctx, codec, NULL)) < 0) {
        fprintf(stderr, "Failed to open codec for stream #%u\n", 233);
        goto init_decoder_fail;
    }

    decoder = (FFDecoder *)malloc(sizeof(FFDecoder));
    decoder -> avc_ctx = avc_ctx;
    decoder -> ctx_opaque = ctx_opaque;
    decoder -> is_hw = is_hw;
    decoder -> hw_pix_fmt = ctx_opaque -> hw_pix_fmt;
    decoder -> hw_device_ctx = hw_device_ctx;
    *ret = decoder;
    return 0;

init_decoder_fail:
    free(ctx_opaque);
    return -1;
}

void ff_free_decoder (FFDecoder * d) {
    av_buffer_unref(&(d -> hw_device_ctx));
    avcodec_free_context(&(d -> avc_ctx));
    free(d -> ctx_opaque);
    free(d);
}


