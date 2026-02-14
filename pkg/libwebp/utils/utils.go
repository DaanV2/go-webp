// Copyright 2012 Google Inc. All Rights Reserved.
//
// Use of this source code is governed by a BSD-style license
// that can be found in the COPYING file in the root of the source
// tree. An additional intellectual property rights grant can be found
// in the file PATENTS. All contributing project authors may
// be found in the AUTHORS file in the root of the source tree.

package utils



func CheckSizeOverflow(size uint64) bool {
	return size == size_t(size)
}

// DaanV2: This aligns a ptr, don't we need to use the aligned ptr?, but maybe its a ptr in an array of bytes
func WEBP_ALIGN(PTR any) {
//   (((uintptr_t)(PTR) + WEBP_ALIGN_CST) & ~(uintptr_t)WEBP_ALIGN_CST)
}

// Returns (int)floor(log2(n)). n must be > 0.
func BitsLog2Floor(n uint32) int  {
  return 31 ^ gcc.Builtin_CLZ(n)
}

// counts the number of trailing zero
func BitsCtz(n uint32) int { return gcc.Builtin_CTZ(n); }

// memcpy() is the safe way of moving potentially unaligned 32b memory.
func WebPMemToUint32(/* const */ ptr *uint8) uint32 {
  var A uint32
  stdlib.MemCpy(&A, ptr, sizeof(A))
  return A
}

func WebPMemToInt32(/* const */ ptr *uint8) int32 {
  return int32(WebPMemToUint32(ptr))
}

func WebPUint32ToMem(/* const */ ptr *uint8, uint32 val) {
  stdlib.MemCpy(ptr, &val, sizeof(val))
}

func WebPInt32ToMem(/* const */ ptr *uint8, val int) {
  WebPUint32ToMem(ptr, uint32(val))
}


// If PRINT_MEM_INFO is defined, extra info (like total memory used, number of
// alloc/free etc) is printed. For debugging/tuning purpose only (it's slow,
// and not multi-thread safe!).
// An interesting alternative is valgrind's 'massif' tool:
//    https://valgrind.org/docs/manual/ms-manual.html
// Here is an example command line:
/*    valgrind --tool=massif --massif-out-file=massif.out \
               --stacks=yes --alloc-fn=WebPSafeMalloc --alloc-fn=WebPSafeCalloc
      ms_print massif.out
*/
// In addition:
// * if PRINT_MEM_TRAFFIC is defined, all the details of the malloc/free cycles
//   are printed.
// * if MALLOC_FAIL_AT is defined, the global environment variable
//   $MALLOC_FAIL_AT is used to simulate a memory error when calloc or malloc
//   is called for the nth time. Example usage:
//   export MALLOC_FAIL_AT=50 && ./examples/cwebp input.png
// * if MALLOC_LIMIT is defined, the global environment variable $MALLOC_LIMIT
//   sets the maximum amount of memory (in bytes) made available to libwebp.
//   This can be used to emulate environment with very limited memory.
//   Example: export MALLOC_LIMIT=64000000 && ./examples/dwebp picture.webp

// #define PRINT_MEM_INFO
// #define PRINT_MEM_TRAFFIC
// #define MALLOC_FAIL_AT
// #define MALLOC_LIMIT


func Increment(/* const */ v []int) {
  if (!exit_registered) {
// #if defined(MALLOC_FAIL_AT)
    {
      var malloc_fail_at_str *byte = getenv("MALLOC_FAIL_AT")
      if (malloc_fail_at_str != nil) {
        countdown_to_fail = atoi(malloc_fail_at_str)
      }
    }
// #endif
// #if defined(MALLOC_LIMIT)
    {
      var malloc_limit_str *byte = getenv("MALLOC_LIMIT")
// #if MALLOC_LIMIT > 1
      mem_limit = uint64(MALLOC_LIMIT)
// #endif
      if (malloc_limit_str != nil) {
        mem_limit = atoi(malloc_limit_str)
      }
    }
// #endif
    _ = countdown_to_fail
    _ = mem_limit
    atexit(PrintMemInfo)
    exit_registered = 1
  }
  v[0]++
}

func AddMem(ptr *void, size uint64 ) {
  if (ptr != nil) {
    // var b *MemBlock = (*MemBlock)malloc(sizeof(*b))
    if b == nil { {abort() }}
    b.next = all_blocks
    all_blocks = b
    b.ptr = ptr
    b.size = size
    total_mem += size
    total_mem_allocated += size
// #if defined(PRINT_MEM_TRAFFIC)
// #if defined(MALLOC_FAIL_AT)
    fprintf(stderr, "fail-count: %5d [mem=%u]\n", num_malloc_calls + num_calloc_calls, uint32(total_mem))
// #else
    fprintf(stderr, "Mem: %u (+%u)\n", uint32(total_mem), uint32(size))
// #endif
// #endif
    if total_mem > high_water_mark { high_water_mark = total_mem }
  }
}

func SubMem(ptr *void) {
  if (ptr != nil) {
    *MemBlock* b = &all_blocks
    // Inefficient search, but that's just for debugging.
    for *b != nil && (*b).ptr != ptr {b = &(*b).next}
    if (*b == nil) {
      fprintf(stderr, "Invalid pointer free! (%p)\n", ptr)
      abort()
    }
    {
      var block *MemBlock = *b
      *b = block.next
      total_mem -= block.size
// #if defined(PRINT_MEM_TRAFFIC)
      fprintf(stderr, "Mem: %u (-%u)\n", uint32(total_mem), uint32(block.size))
// #endif
      free(block)
    }
  }
}

func CheckSizeOverflow(size uint64 ) bool {
  return size == uint64(size)
}

// Returns 0 in case of overflow of nmemb * size.
func CheckSizeArgumentsOverflow(nmemb uint64, size uint64 ) int {
  total_size := nmemb * size
  if nmemb == 0 { return 1  }
  if uint64(size)> WEBP_MAX_ALLOCABLE_MEMORY / nmemb { return 0  }
  if !CheckSizeOverflow(total_size) { return 0  }
// #if defined(PRINT_MEM_INFO) && defined(MALLOC_FAIL_AT)
  if (countdown_to_fail > 0 && --countdown_to_fail == 0) {
    return 0;  // fake fail!
  }
// #endif
// #if defined(PRINT_MEM_INFO) && defined(MALLOC_LIMIT)
  if (mem_limit > 0) {
    new_total_mem := uint64(total_mem)+ total_size
    if (!CheckSizeOverflow(new_total_mem) || new_total_mem > mem_limit) {
      return 0;  // fake fail!
    }
  }
// #endif

  return 1
}

// Copy width x height pixels from 'src' to 'dst' honoring the strides.
func WebPCopyPlane(/* const */ src *uint8, src_stride int, dst []uint8, dst_stride int, width, height int) {
  assert.Assert(src != nil && dst != nil)
  assert.Assert(abs(src_stride) >= width && abs(dst_stride) >= width)
  for (height-- > 0) {
    stdlib.MemCpy(dst, src, width)
    src += src_stride
    dst += dst_stride
  }
}

// Copy ARGB pixels from 'src' to 'dst' honoring strides. 'src' and 'dst' are
// assumed to be already allocated and using ARGB data.
func WebPCopyPixels(/* const */ src *picture.Picture, /*const*/ dst *picture.Picture) {
  assert.Assert(src != nil && dst != nil)
  assert.Assert(src.width == dst.width && src.height == dst.height)
  assert.Assert(src.use_argb && dst.use_argb)
  WebPCopyPlane((*uint8)src.argb, 4 * src.argb_stride, (*uint8)dst.argb, 4 * dst.argb_stride, 4 * src.width, src.height)
}


// Returns count of unique colors in 'pic', assuming pic.use_argb is true.
// If the unique color count is more than constants.MAX_PALETTE_SIZE, returns
// constants.MAX_PALETTE_SIZE+1.
// If 'palette' is not nil and number of unique colors is less than or equal to
// constants.MAX_PALETTE_SIZE, also outputs the actual unique colors into 'palette'.
// Note: 'palette' is assumed to be an array already allocated with at least
// constants.MAX_PALETTE_SIZE elements.
// TODO(vrabaud) remove whenever we can break the ABI.
func WebPGetColorPalette(/* const  */pic *picture.Picture, /*const*/  palette []uint32/* (constants.MAX_PALETTE_SIZE) */) int {
  return GetColorPalette(pic, palette)
}