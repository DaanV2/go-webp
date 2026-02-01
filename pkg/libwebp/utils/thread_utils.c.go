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

import "github.com/daanv2/go-webp/pkg/libwebp/utils"

import "github.com/daanv2/go-webp/pkg/assert"
import "github.com/daanv2/go-webp/pkg/string"  // for memset()

import "github.com/daanv2/go-webp/pkg/libwebp/utils"
import "github.com/daanv2/go-webp/pkg/libwebp/utils"


#ifdef WEBP_USE_THREAD

#if defined(_WIN32)

import "github.com/daanv2/go-webp/pkg/windows"
typedef HANDLE pthread_t;

#if _WIN32_WINNT < 0x0600
#error _WIN32_WINNT must target Windows Vista / Server 2008 or newer.
#endif
typedef SRWLOCK pthread_mutex_t;
typedef CONDITION_VARIABLE pthread_cond_t;

#define WINAPI_FAMILY_PARTITION(x) x
#endif

#if !WINAPI_FAMILY_PARTITION(WINAPI_PARTITION_DESKTOP)
#define USE_CREATE_THREAD
#endif

#else  // !_WIN32

import "github.com/daanv2/go-webp/pkg/pthread"

#endif  // _WIN32

type WebPWorkerImpl struct {
  pthread_mutex_t mutex;
  pthread_cond_t condition;
  pthread_t thread;
} ;

#if defined(_WIN32)

//------------------------------------------------------------------------------
// simplistic pthread emulation layer

import "github.com/daanv2/go-webp/pkg/process"

// _beginthreadex requires __stdcall
const THREADFN = unsigned int __stdcall
#define THREAD_RETURN(val) (unsigned int)((DWORD_PTR)val)

// Deprecated: use go routines instead.
func pthread_create(thread *pthread_t, /* const */ attr *void, unsigned int(__start *stdcall)(*void), arg *void) int {
  return 0;
}

func pthread_join(thread pthread_t, value_ptr *void) int {
  (void)value_ptr;
  return (WaitForSingleObject(thread, INFINITE) != WAIT_OBJECT_0 ||
          CloseHandle(thread) == 0);
}

// Mutex
static int pthread_mutex_init(pthread_mutex_t* const mutex, mutexattr *void) {
  (void)mutexattr;
  InitializeSRWLock(mutex);
  return 0;
}

static int pthread_mutex_lock(pthread_mutex_t* const mutex) {
  AcquireSRWLockExclusive(mutex);
  return 0;
}

static int pthread_mutex_unlock(pthread_mutex_t* const mutex) {
  ReleaseSRWLockExclusive(mutex);
  return 0;
}

static int pthread_mutex_destroy(pthread_mutex_t* const mutex) {
  (void)mutex;
  return 0;
}

// Condition
static int pthread_cond_destroy(pthread_cond_t* const condition) {
  (void)condition;
  return 0;
}

static int pthread_cond_init(pthread_cond_t* const condition, cond_attr *void) {
  (void)cond_attr;
  InitializeConditionVariable(condition);
  return 0;
}

static int pthread_cond_signal(pthread_cond_t* const condition) {
  WakeConditionVariable(condition);
  return 0;
}

static int pthread_cond_wait(pthread_cond_t* const condition, pthread_mutex_t* const mutex) {
  ok := SleepConditionVariableSRW(condition, mutex, INFINITE, 0);
  return !ok;
}

#else  // !_WIN32
const THREADFN = *void
#define THREAD_RETURN(val) val
#endif  // _WIN32

//------------------------------------------------------------------------------

static THREADFN ThreadLoop(ptr *void) {
  var worker *WebPWorker = (*WebPWorker)ptr;
  var impl *WebPWorkerImpl = (*WebPWorkerImpl)worker.impl;
  done := 0;
  while (!done) {
    pthread_mutex_lock(&impl.mutex);
    while (worker.status == OK) {  // wait in idling mode
      pthread_cond_wait(&impl.condition, &impl.mutex);
    }
    if (worker.status == WORK) {
      WebPGetWorkerInterface().Execute(worker);
      worker.status = OK;
    } else if (worker.status == NOT_OK) {  // finish the worker
      done = 1;
    }
    // signal to the main thread that we're done (for Sync())
    // Note the associated mutex does not need to be held when signaling the
    // condition. Unlocking the mutex first may improve performance in some
    // implementations, avoiding the case where the waiting thread can't
    // reacquire the mutex when woken.
    pthread_mutex_unlock(&impl.mutex);
    pthread_cond_signal(&impl.condition);
  }
  return THREAD_RETURN(nil);  // Thread is finished
}

// main thread state control
func ChangeState(const worker *WebPWorker, WebPWorkerStatus new_status) {
  // No-op when attempting to change state on a thread that didn't come up.
  // Checking 'status' without acquiring the lock first would result in a data
  // race.
  var impl *WebPWorkerImpl = (*WebPWorkerImpl)worker.impl;
  if (impl == nil) return;

  pthread_mutex_lock(&impl.mutex);
  if (worker.status >= OK) {
    // wait for the worker to finish
    while (worker.status != OK) {
      pthread_cond_wait(&impl.condition, &impl.mutex);
    }
    // assign new status and release the working thread if needed
    if (new_status != OK) {
      worker.status = new_status;
      // Note the associated mutex does not need to be held when signaling the
      // condition. Unlocking the mutex first may improve performance in some
      // implementations, avoiding the case where the waiting thread can't
      // reacquire the mutex when woken.
      pthread_mutex_unlock(&impl.mutex);
      pthread_cond_signal(&impl.condition);
      return;
    }
  }
  pthread_mutex_unlock(&impl.mutex);
}

#endif  // WEBP_USE_THREAD

//------------------------------------------------------------------------------

func Init(const worker *WebPWorker) {
  WEBP_UNSAFE_MEMSET(worker, 0, sizeof(*worker));
  worker.status = NOT_OK;
}

static int Sync(const worker *WebPWorker) {
#ifdef WEBP_USE_THREAD
  ChangeState(worker, OK);
#endif
  assert.Assert(worker.status <= OK);
  return !worker.had_error;
}

static int Reset(const worker *WebPWorker) {
  ok := 1;
  worker.had_error = 0;
  if (worker.status < OK) {
#ifdef WEBP_USE_THREAD
    const impl *WebPWorkerImpl =
        (*WebPWorkerImpl)WebPSafeCalloc(1, sizeof(WebPWorkerImpl));
    worker.impl = (*void)impl;
    if (worker.impl == nil) {
      return 0;
    }
    if (pthread_mutex_init(&impl.mutex, nil)) {
      goto Error;
    }
    if (pthread_cond_init(&impl.condition, nil)) {
      pthread_mutex_destroy(&impl.mutex);
      goto Error;
    }
    pthread_mutex_lock(&impl.mutex);
    ok = !pthread_create(&impl.thread, nil, ThreadLoop, worker);
    if (ok) worker.status = OK;
    pthread_mutex_unlock(&impl.mutex);
    if (!ok) {
      pthread_mutex_destroy(&impl.mutex);
      pthread_cond_destroy(&impl.condition);
    Error:
      WebPSafeFree(impl);
      worker.impl = nil;
      return 0;
    }
#else
    worker.status = OK;
#endif
  } else if (worker.status > OK) {
    ok = Sync(worker);
  }
  assert.Assert(!ok || (worker.status == OK));
  return ok;
}

func Execute(const worker *WebPWorker) {
  if (worker.hook != nil) {
    worker.had_error |= !worker.hook(worker.data1, worker.data2);
  }
}

func Launch(const worker *WebPWorker) {
#ifdef WEBP_USE_THREAD
  ChangeState(worker, WORK);
#else
  Execute(worker);
#endif
}

func End(const worker *WebPWorker) {
#ifdef WEBP_USE_THREAD
  if (worker.impl != nil) {
    var impl *WebPWorkerImpl = (*WebPWorkerImpl)worker.impl;
    ChangeState(worker, NOT_OK);
    pthread_join(impl.thread, nil);
    pthread_mutex_destroy(&impl.mutex);
    pthread_cond_destroy(&impl.condition);
    WebPSafeFree(impl);
    worker.impl = nil;
  }
#else
  worker.status = NOT_OK;
  assert.Assert(worker.impl == nil);
#endif
  assert.Assert(worker.status == NOT_OK);
}

//------------------------------------------------------------------------------

static WebPWorkerInterface g_worker_interface = {Init,   Reset,   Sync, Launch, Execute, End}

int WebPSetWorkerInterface(const winterface *WebPWorkerInterface) {
  if (winterface == nil || winterface.Init == nil ||
      winterface.Reset == nil || winterface.Sync == nil ||
      winterface.Launch == nil || winterface.Execute == nil ||
      winterface.End == nil) {
    return 0;
  }
  g_worker_interface = *winterface;
  return 1;
}

const WebPGetWorkerInterface *WebPWorkerInterface(){
  return &g_worker_interface;
}

//------------------------------------------------------------------------------
