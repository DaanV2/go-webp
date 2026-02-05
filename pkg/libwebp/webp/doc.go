// This API allows streamlined decoding of partial data.
// Picture can be incrementally decoded as data become available thanks to the
// WebPIDecoder object. This object can be left in a SUSPENDED state if the
// picture is only partially decoded, pending additional input.
// Code example:
//
//	WebPInitDecBuffer(&output_buffer);
//	output_buffer.colorspace = mode;
//	...
//	idec *WebPIDecoder = WebPINewDecoder(&output_buffer);
//	while (additional_data_is_available) {
//		// ... (get additional data in some new_data[] buffer)
//		status = WebPIAppend(idec, new_data, new_data_size);
//		if (status != VP8_STATUS_OK && status != VP8_STATUS_SUSPENDED) {
//		break;    // an error occurred.
//		}
//
//		// The above call decodes the current available buffer.
//		// Part of the image can now be refreshed by calling
//		// WebPIDecGetRGB()/WebPIDecGetYUVA() etc.
//	}
//
// Code sample for using the advanced decoding API
//
//	// A) Init a configuration object
//	WebPDecoderConfig config;
//	CHECK(WebPInitDecoderConfig(&config));
//
//	// B) optional: retrieve the bitstream's features.
//	CHECK(WebPGetFeatures(data, data_size, &config.input) == VP8_STATUS_OK);
//
//	// C) Adjust 'config', if needed
//	config.options.no_fancy_upsampling = 1;
//	config.output.colorspace = MODE_BGRA;
//	// etc.
//
//	// Note that you can also make config.output point to an externally
//	// supplied memory buffer, provided it's big enough to store the decoded
//	// picture. Otherwise, config.output will just be used to allocate memory
//	// and store the decoded picture.
//
//	// D) Decode!
//	CHECK(WebPDecode(data, data_size, &config) == VP8_STATUS_OK);
//
//	// E) Decoded image is now in config.output (and config.output.u.RGBA)
//
//	// F) Reclaim memory allocated in config's object. It's safe to call
//	// this function even if the memory is external and wasn't allocated
//	// by WebPDecode().
//	WebPFreeDecBuffer(&config.output);
package webp
