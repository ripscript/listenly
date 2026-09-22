from grpc_server import serve
import os

if __name__ == "__main__":
    serve(os.getenv("GRPC_PORT", "50054"))