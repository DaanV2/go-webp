# Webp Specification

This is a summary of google [documentation](https://developers.google.com/speed/webp/docs/riff_container).
It was written to summarize, and as wel as a learning tool.



## Spec

Webp is an image format. Where by its data formatted in files / data stream via Riff container. Raw data can be compressed using lossy VP8 or WebP lossless encoding.

Webp aim was to be faster of the network. While holding parity such as animation, color profile and metadata.
Metadata is either stored in `Exif` or `XMP` format.

### File Format

#### WebP Fileheader

Each file should start with a [FourCC](#fourcc) with value "RIFF"

### Metadata Format

### Riff Format

Riff uses block of bytes and strings it together.
Each chunk has the following format:

- 4 bytes for the [FourCC](#fourcc)
- 4 bytes as uint32 for Size
- Payload is the collection of bytes representing the data, if the Size is odd, pad it with a byte with value 0


#### FourCC

Means: Four Character Code, or the identification of the chunk