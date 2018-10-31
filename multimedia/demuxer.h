//
// Created by tigerjang on 2018/10/30.
//

#ifndef DSHREPOSITORY_DEMUXER_H
#define DSHREPOSITORY_DEMUXER_H

#include <stdio.h>

#include <libavcodec/avcodec.h>
#include <libavformat/avformat.h>
//#include <libavutil/pixdesc.h>
#include <libavutil/hwcontext.h>
#include <libavutil/opt.h>


#define NUM_MEDIA_TYPE 6

typedef struct FFIOContext {
} FFIOContext;

typedef struct FFDemuxer {
//    unsigned int n_stream;
    AVFormatContext * avf_ctx;
    int best_streams[NUM_MEDIA_TYPE];
} FFDemuxer;

int ff_new_demuxer (FFDemuxer ** ret, const char * uri, FFIOContext * io_ctx);
void ff_free_demuxer (FFDemuxer * d);
int ff_demuxer_get_codec_paras (FFDemuxer * d, AVCodecParameters ** ret, unsigned int stream_id);

typedef struct FFDecoder {
    AVCodecContext * avc_ctx;
} FFDecoder;

int ff_new_decoder (FFDecoder ** ret, AVCodecParameters * codec_paras);
void ff_free_decoder (FFDecoder * d);


static const int GO_AVERROR(int e)
{
  return AVERROR(e);
}

#endif //DSHREPOSITORY_DEMUXER_H

