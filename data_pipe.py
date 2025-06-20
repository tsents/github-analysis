import zlib
import argparse

CHUNK_SIZE = 1024
OPEN_BRACKET = b"{"
CLOSE_BRACKET = b"}"

def decompress_stream():
    return zlib.decompressobj(16 + zlib.MAX_WBITS)

def pipe_data(compressed_data_file):
    decompressor = decompress_stream()
    unused_data = b""
    while True:
        unused_data += decompressor.decompress(compressed_data_file.read(CHUNK_SIZE))
        if not unused_data:
            break
        try:
            unused_data = process_chunk(unused_data)
        except:
            pass
    return


def process_chunk(raw_data) -> bytes:
    print(raw_data.decode())
    if (raw_data[0] != OPEN_BRACKET):
        raise Exception(f"Data not starting with {OPEN_BRACKET}")
    indent_depth = 0
    for i in raw_data:
        if raw_data[i] == OPEN_BRACKET:
            indent_depth += 1
        if raw_data[i] == CLOSE_BRACKET:
            indent_depth -= 1
        if indent_depth == 0:
            print(raw_data[:i+1])
            return raw_data[i+1:]
    return b""

def main():
    parser = argparse.ArgumentParser(prog='data_pipe');
    parser.add_argument('file', help="json.gz file to parse", type=argparse.FileType('rb'));
    args = parser.parse_args()

    pipe_data(args.file);

    args.file.close();

if __name__ == '__main__':
    main()
