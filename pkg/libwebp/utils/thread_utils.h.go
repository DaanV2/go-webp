package utils

// Copyright 2011 Google Inc. All Rights Reserved.
//
// Use of this source code is governed by a BSD-style license
// that can be found in the COPYING file in the root of the source
// tree. An additional intellectual property rights grant can be found
// in the file PATENTS. All contributing project authors may
// be found in the AUTHORS file in the root of the source tree.
// -----------------------------------------------------------------------------
//
// Multi-threaded worker
//
// Author: Skal (pascal.massimino@gmail.com)


#ifdef HAVE_CONFIG_H
import "github.com/daanv2/go-webp/pkg/libwebp/webp"
#endif

import "github.com/daanv2/go-webp/pkg/libwebp/utils"
import "github.com/daanv2/go-webp/pkg/libwebp/webp"

WEBP_ASSUME_UNSAFE_INDEXABLE_ABI

#ifdef __cplusplus
extern "C" {
#endif

// State of the worker thread object
type WebPWorkerStatus int

const (
  NOT_OK WebPWorkerStatus = iota  // object is unusable
  OK          // ready to work
  WORK         // busy finishing the current task
}

// Function to be called by the worker thread. Takes two opaque pointers as
// arguments (data1 and data2), and should return false in case of error.
typedef int (*WebPWorkerHook)(*void, *void);

// Synchronization object used to launch job in the worker thread
type <Foo> struct {
  impl *void;  // platform-dependent implementation worker details
  WebPWorkerStatus status;
  WebPWorkerHook hook;  // hook to call
  data *void1;          // first argument passed to 'hook'
  data *void2;          // second argument passed to 'hook'
  int had_error;        // return value of the last call to 'hook'
} WebPWorker;

// The interface for all thread-worker related functions. All these functions
// must be implemented.
type <Foo> struct {
  // Must be called first, before any other method.
  func (*Init)(const worker *WebPWorker);
  // Must be called to initialize the object and spawn the thread. Re-entrant.
  // Will potentially launch the thread. Returns false in case of error.
  int (*Reset)(const worker *WebPWorker);
  // Makes sure the previous work is finished. Returns true if worker.had_error
  // was not set and no error condition was triggered by the working thread.
  int (*Sync)(const worker *WebPWorker);
  // Triggers the thread to call hook() with data1 and data2 arguments. These
  // hook/data1/data2 values can be changed at any time before calling this
  // function, but not be changed afterward until the next call to Sync().
  func (*Launch)(const worker *WebPWorker);
  // This function is similar to Launch() except that it calls the
  // hook directly instead of using a thread. Convenient to bypass the thread
  // mechanism while still using the WebPWorker structs. Sync() must
  // still be called afterward (for error reporting).
  func (*Execute)(const worker *WebPWorker);
  // Kill the thread and terminate the object. To use the object again, one
  // must call Reset() again.
  func (*End)(const worker *WebPWorker);
} WebPWorkerInterface;

// Install a new set of threading functions, overriding the defaults. This
// should be done before any workers are started, i.e., before any encoding or
// decoding takes place. The contents of the interface struct are copied, it
// is safe to free the corresponding memory after this call. This function is
// not thread-safe. Return false in case of invalid pointer or methods.
 int WebPSetWorkerInterface(
    const winterface *WebPWorkerInterface);

// Retrieve the currently set thread worker interface.
 const WebPGetWorkerInterface *WebPWorkerInterface(void);

//------------------------------------------------------------------------------

#ifdef __cplusplus
}  // extern "C"
#endif

#endif  // WEBP_UTILS_THREAD_UTILS_H_
