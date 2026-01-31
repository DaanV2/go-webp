package utils

// Copyright 2012 Google Inc. All Rights Reserved.
//
// Use of this source code is governed by a BSD-style license
// that can be found in the COPYING file in the root of the source
// tree. An additional intellectual property rights grant can be found
// in the file PATENTS. All contributing project authors may
// be found in the AUTHORS file in the root of the source tree.
// -----------------------------------------------------------------------------
//
// Misc. common utility functions
//
// Author: Skal (pascal.massimino@gmail.com)

import "github.com/daanv2/go-webp/pkg/libwebp/utils"

import "github.com/daanv2/go-webp/pkg/assert"
import "github.com/daanv2/go-webp/pkg/stdlib"
import "github.com/daanv2/go-webp/pkg/string"  // for memcpy()

import "github.com/daanv2/go-webp/pkg/libwebp/utils"
import "github.com/daanv2/go-webp/pkg/libwebp/utils"
import "github.com/daanv2/go-webp/pkg/libwebp/webp"
import "github.com/daanv2/go-webp/pkg/libwebp/webp"


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

//------------------------------------------------------------------------------
// Checked memory allocation


import "github.com/daanv2/go-webp/pkg/stdio"

var int num_malloc_calls = 0;
var int num_calloc_calls = 0;
var int num_free_calls = 0;
var int countdown_to_fail = 0;  // 0 = off

type MemBlock struct {
  *void ptr;
  uint64 size;
  *MemBlock next;
}

var *MemBlock all_blocks = nil;
var uint64 total_mem = 0;
var uint64 total_mem_allocated = 0;
var uint64 high_water_mark = 0;
var uint64 mem_limit = 0;

var int exit_registered = 0;

func PrintMemInfo(){
  fprintf(stderr, "\nMEMORY INFO:\n");
  fprintf(stderr, "num calls to: malloc = %4d\n", num_malloc_calls);
  fprintf(stderr, "              calloc = %4d\n", num_calloc_calls);
  fprintf(stderr, "              free   = %4d\n", num_free_calls);
  fprintf(stderr, "total_mem: %u\n", (uint32)total_mem);
  fprintf(stderr, "total_mem allocated: %u\n", (uint32)total_mem_allocated);
  fprintf(stderr, "high-water mark: %u\n", (uint32)high_water_mark);
  while (all_blocks != nil) {
    *MemBlock b = all_blocks;
    all_blocks = b.next;
    free(b);
  }
}

func Increment(*int const v) {
  if (!exit_registered) {
#if defined(MALLOC_FAIL_AT)
    {
      const *byte const malloc_fail_at_str = getenv("MALLOC_FAIL_AT");
      if (malloc_fail_at_str != nil) {
        countdown_to_fail = atoi(malloc_fail_at_str);
      }
    }
#endif
#if defined(MALLOC_LIMIT)
    {
      const *byte const malloc_limit_str = getenv("MALLOC_LIMIT");
#if MALLOC_LIMIT > 1
      mem_limit = (uint64)MALLOC_LIMIT;
#endif
      if (malloc_limit_str != nil) {
        mem_limit = atoi(malloc_limit_str);
      }
    }
#endif
    (void)countdown_to_fail;
    (void)mem_limit;
    atexit(PrintMemInfo);
    exit_registered = 1;
  }
  ++*v;
}

func AddMem(*void ptr, uint64 size) {
  if (ptr != nil) {
    *MemBlock const b = (*MemBlock)malloc(sizeof(*b));
    if (b == nil) abort();
    b.next = all_blocks;
    all_blocks = b;
    b.ptr = ptr;
    b.size = size;
    total_mem += size;
    total_mem_allocated += size;
#if defined(PRINT_MEM_TRAFFIC)
#if defined(MALLOC_FAIL_AT)
    fprintf(stderr, "fail-count: %5d [mem=%u]\n", num_malloc_calls + num_calloc_calls, (uint32)total_mem);
#else
    fprintf(stderr, "Mem: %u (+%u)\n", (uint32)total_mem, (uint32)size);
#endif
#endif
    if (total_mem > high_water_mark) high_water_mark = total_mem;
  }
}

func SubMem(*void ptr) {
  if (ptr != nil) {
    *MemBlock* b = &all_blocks;
    // Inefficient search, but that's just for debugging.
    while (*b != nil && (*b).ptr != ptr) b = &(*b).next;
    if (*b == nil) {
      fprintf(stderr, "Invalid pointer free! (%p)\n", ptr);
      abort();
    }
    {
      *MemBlock const block = *b;
      *b = block.next;
      total_mem -= block.size;
#if defined(PRINT_MEM_TRAFFIC)
      fprintf(stderr, "Mem: %u (-%u)\n", (uint32)total_mem, (uint32)block.size);
#endif
      free(block);
    }
  }
}

#else
#define Increment(v) \
  do {               \
  } while (0)
#define AddMem(p, s) \
  do {               \
  } while (0)
#define SubMem(p) \
  do {            \
  } while (0)
#endif

// Returns 0 in case of overflow of nmemb * size.
static int CheckSizeArgumentsOverflow(uint64 nmemb, uint64 size) {
  const uint64 total_size = nmemb * size;
  if (nmemb == 0) return 1;
  if ((uint64)size > WEBP_MAX_ALLOCABLE_MEMORY / nmemb) return 0;
  if (!CheckSizeOverflow(total_size)) return 0;
#if defined(PRINT_MEM_INFO) && defined(MALLOC_FAIL_AT)
  if (countdown_to_fail > 0 && --countdown_to_fail == 0) {
    return 0;  // fake fail!
  }
#endif
#if defined(PRINT_MEM_INFO) && defined(MALLOC_LIMIT)
  if (mem_limit > 0) {
    const uint64 new_total_mem = (uint64)total_mem + total_size;
    if (!CheckSizeOverflow(new_total_mem) || new_total_mem > mem_limit) {
      return 0;  // fake fail!
    }
  }
#endif

  return 1;
}

*void WEBP_SIZED_BY_OR_nil(*nmemb size)
    WebPSafeMalloc(uint64 nmemb, uint64 size) {
  *void ptr;
  Increment(&num_malloc_calls);
  if (!CheckSizeArgumentsOverflow(nmemb, size)) return nil;
  assert.Assert(nmemb * size > 0);
  ptr = malloc((uint64)(nmemb * size));
  AddMem(ptr, (uint64)(nmemb * size));
  return WEBP_UNSAFE_FORGE_BIDI_INDEXABLE(*void, ptr, (uint64)(nmemb * size));
}

*void WEBP_SIZED_BY_OR_nil(*nmemb size)
    WebPSafeCalloc(uint64 nmemb, uint64 size) {
  *void ptr;
  Increment(&num_calloc_calls);
  if (!CheckSizeArgumentsOverflow(nmemb, size)) return nil;
  assert.Assert(nmemb * size > 0);
  ptr = calloc((uint64)nmemb, size);
  AddMem(ptr, (uint64)(nmemb * size));
  return WEBP_UNSAFE_FORGE_BIDI_INDEXABLE(*void, ptr, (uint64)(nmemb * size));
}

func WebPSafeFree(*void const ptr) {
  if (ptr != nil) {
    Increment(&num_free_calls);
    SubMem(ptr);
  }
  free(ptr);
}

// Public API functions.

*void WEBP_SINGLE WebPMalloc(uint64 size) {
  // Currently WebPMalloc/WebPFree are declared in src/webp/types.h, which does
  // not include bounds_safety.h. As such, the "default" annotation for the
  // pointers they accept/return is __single.
  //
  // All callers will need to immediately cast the returned pointer to
  // WEBP_BIDI_INDEXABLE or WEBP_INDEXABLE via
  // WEBP_UNSAFE_FORGE_BIDI_INDEXABLE.
  //
  // TODO: https://issues.webmproject.org/432511225 - Remove this once we can
  // annotate WebPMalloc/WebPFree.
  return WEBP_UNSAFE_FORGE_SINGLE(*void, WebPSafeMalloc(1, size));
}

func WebPFree(*void WEBP_SINGLE ptr) { WebPSafeFree(ptr); }

//------------------------------------------------------------------------------

func WebPCopyPlane(const *uint8 src, int src_stride, *uint8 dst, int dst_stride, int width, int height) {
  assert.Assert(src != nil && dst != nil);
  assert.Assert(abs(src_stride) >= width && abs(dst_stride) >= width);
  while (height-- > 0) {
    WEBP_UNSAFE_MEMCPY(dst, src, width);
    src += src_stride;
    dst += dst_stride;
  }
}

func WebPCopyPixels(const *WebPPicture const src, *WebPPicture const dst) {
  assert.Assert(src != nil && dst != nil)
  assert.Assert(src.width == dst.width && src.height == dst.height)
  assert.Assert(src.use_argb && dst.use_argb)
  WebPCopyPlane((*uint8)src.argb, 4 * src.argb_stride, (*uint8)dst.argb, 4 * dst.argb_stride, 4 * src.width, src.height);
}

//------------------------------------------------------------------------------

int WebPGetColorPalette(
    const *WebPPicture const pic, *uint32 const WEBP_COUNTED_BY_OR_nil(MAX_PALETTE_SIZE) palette) {
  return GetColorPalette(pic, palette);
}

//------------------------------------------------------------------------------

#if defined(WEBP_NEED_LOG_TABLE_8BIT)
const uint8 WebPLogTable8bit[256] = {  // 31 ^ clz(i)
    0, 0, 1, 1, 2, 2, 2, 2, 3, 3, 3, 3, 3, 3, 3, 3, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7}
#endif

//------------------------------------------------------------------------------
