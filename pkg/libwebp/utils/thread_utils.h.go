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

// State of the worker thread object
type WebPWorkerStatus int

const (
  NOT_OK WebPWorkerStatus = iota  // object is unusable
  OK          // ready to work
  WORK         // busy finishing the current task
)

// Function to be called by the worker thread. Takes two opaque pointers as
// arguments (data1 and data2), and should return false in case of error.
type WebPWorkerHook = func(*void, *void)int

// Synchronization object used to launch job in the worker thread
type WebPWorker struct {
  impl *void  // platform-dependent implementation worker details
  status WebPWorkerStatus
  hook WebPWorkerHook  // hook to call
  data1 *void          // first argument passed to 'hook'
  data2 *void          // second argument passed to 'hook'
  int had_error        // return value of the last call to 'hook'
}

// The interface for all thread-worker related functions. All these functions
// must be implemented.
type WebPWorkerInterface struct {
  // Must be called first, before any other method.
  Init func(/* const */ worker *WebPWorker)
  // Must be called to initialize the object and spawn the thread. Re-entrant.
  // Will potentially launch the thread. Returns false in case of error.
  Reset func(/* const */ worker *WebPWorker)int
  // Makes sure the previous work is finished. Returns true if worker.had_error
  // was not set and no error condition was triggered by the working thread.
  Sync func(/* const */ worker *WebPWorker)int
  // Triggers the thread to call hook() with data1 and data2 arguments. These
  // hook/data1/data2 values can be changed at any time before calling this
  // function, but not be changed afterward until the next call to Sync().
  Launch func(/* const */ worker *WebPWorker)
  // This function is similar to Launch() except that it calls the
  // hook directly instead of using a thread. Convenient to bypass the thread
  // mechanism while still using the WebPWorker structs. Sync() must
  // still be called afterward (for error reporting).
  Execute func(/* const */ worker *WebPWorker)
  // Kill the thread and terminate the object. To use the object again, one
  // must call Reset() again.
  End func(/* const */ worker *WebPWorker)
} 

// Install a new set of threading functions, overriding the defaults. This
// should be done before any workers are started, i.e., before any encoding or
// decoding takes place. The contents of the interface struct are copied, it
// is safe to free the corresponding memory after this call. This function is
// not thread-safe. Return false in case of invalid pointer or methods.
func WebPSetWorkerInterface(/* const */ winterface *WebPWorkerInterface) int {
	// TODO: implement this function
	return 0
}

// Retrieve the currently set thread worker interface.
func WebPWorkerInterface() *WebPGetWorkerInterface {
	// TODO: implement this function
	return nil
}


