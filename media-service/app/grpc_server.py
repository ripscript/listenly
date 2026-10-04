import time
import grpc
from concurrent import futures

from media.v1 import media_pb2, media_pb2_grpc


from media_extractor import (
    get_media_info,
    get_audio_stream_url,
    search_youtube,
    AUDIO_QUALITY_MAP,
    MediaExtractionError,
)


class MediaServiceServicer(media_pb2_grpc.MediaServiceServicer):
    def GetMediaInfo(self, request, context):
        try:
            info = get_media_info(request.youtube_url)
        except MediaExtractionError as e:
            context.set_code(grpc.StatusCode.INVALID_ARGUMENT)
            context.set_details(str(e))
            return media_pb2.GetMediaInfoResponse()

        return media_pb2.GetMediaInfoResponse(
            video_id=info["video_id"],
            title=info["title"],
            channel=info["channel"],
            duration_seconds=info["duration_seconds"],
            thumbnail_url=info["thumbnail_url"],
        )

    def GetStreamUrl(self, request, context):
        quality = AUDIO_QUALITY_MAP.get(request.quality, AUDIO_QUALITY_MAP[0])
        try:
            stream_url = get_audio_stream_url(request.video_id, quality)
        except MediaExtractionError as e:
            context.set_code(grpc.StatusCode.NOT_FOUND)
            context.set_details(str(e))
            return media_pb2.GetStreamUrlResponse()

        # expires_at informational saja, client TETAP wajib minta ulang saat butuh
        expires_at = time.strftime("%Y-%m-%dT%H:%M:%SZ", time.gmtime(time.time() + 3600 * 5))

        return media_pb2.GetStreamUrlResponse(
            stream_url=stream_url,
            expires_at=expires_at,
        )

    def SearchYouTube(self, request, context):
        try:
            results = search_youtube(request.query, request.limit or 10)
        except MediaExtractionError as e:
            context.set_code(grpc.StatusCode.INVALID_ARGUMENT)
            context.set_details(str(e))
            return media_pb2.SearchYouTubeResponse()

        return media_pb2.SearchYouTubeResponse(
            results=[
                media_pb2.SearchYouTubeResult(
                    video_id=r["video_id"],
                    title=r["title"],
                    channel=r["channel"],
                    duration_seconds=r["duration_seconds"],
                    thumbnail_url=r["thumbnail_url"],
                )
                for r in results
            ]
        )


def recovery_interceptor():
    """Setara pkg/grpcserver/recovery.go di Go — cegah 1 panic mematikan seluruh server."""
    class RecoveryInterceptor(grpc.ServerInterceptor):
        def intercept_service(self, continuation, handler_call_details):
            handler = continuation(handler_call_details)
            if handler is None:
                return handler

            def wrap_unary(behavior):
                def wrapper(request, context):
                    try:
                        return behavior(request, context)
                    except Exception as e:
                        context.set_code(grpc.StatusCode.INTERNAL)
                        context.set_details(f"internal server error: {e}")
                        return None
                return wrapper

            if handler.unary_unary:
                return grpc.unary_unary_rpc_method_handler(
                    wrap_unary(handler.unary_unary),
                    request_deserializer=handler.request_deserializer,
                    response_serializer=handler.response_serializer,
                )
            return handler

    return RecoveryInterceptor()


def serve(port: str = "50054"):
    server = grpc.server(
        futures.ThreadPoolExecutor(max_workers=10),
        interceptors=[recovery_interceptor()],
    )
    media_pb2_grpc.add_MediaServiceServicer_to_server(MediaServiceServicer(), server)
    server.add_insecure_port(f"[::]:{port}")
    server.start()
    print(f"media-service (gRPC) listening on port {port}")
    server.wait_for_termination()


if __name__ == "__main__":
    import os
    serve(os.getenv("GRPC_PORT", "50054"))