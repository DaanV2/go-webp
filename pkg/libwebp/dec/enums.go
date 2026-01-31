package dec

// Copyright 2015 Google Inc. All Rights Reserved.
//
// Use of this source code is governed by a BSD-style license
// that can be found in the COPYING file in the root of the source
// tree. An additional intellectual property rights grant can be found
// in the file PATENTS. All contributing project authors may
// be found in the AUTHORS file in the root of the source tree.

type PREDICTION_MODE int

const (
	B_DC_PRED PREDICTION_MODE = 0
	// 4x4 modes
	B_TM_PRED  PREDICTION_MODE = 1
	B_VE_PRED  PREDICTION_MODE = 2
	B_HE_PRED  PREDICTION_MODE = 3
	B_RD_PRED  PREDICTION_MODE = 4
	B_VR_PRED  PREDICTION_MODE = 5
	B_LD_PRED  PREDICTION_MODE = 6
	B_VL_PRED  PREDICTION_MODE = 7
	B_HD_PRED  PREDICTION_MODE = 8
	B_HU_PRED  PREDICTION_MODE = 9
	NUM_BMODES PREDICTION_MODE = B_HU_PRED + 1 - B_DC_PRED
	// = 10

	// Luma16 or UV modes
	DC_PRED PREDICTION_MODE = B_DC_PRED
	V_PRED  PREDICTION_MODE = B_VE_PRED
	H_PRED  PREDICTION_MODE = B_HE_PRED
	TM_PRED PREDICTION_MODE = B_TM_PRED
	B_PRED  PREDICTION_MODE = NUM_BMODES
	// refined I4x4 mode
	NUM_PRED_MODES PREDICTION_MODE = 4

	// special modes
	B_DC_PRED_NOTOP     PREDICTION_MODE = 4
	B_DC_PRED_NOLEFT    PREDICTION_MODE = 5
	B_DC_PRED_NOTOPLEFT PREDICTION_MODE = 6
	NUM_B_DC_MODES      PREDICTION_MODE = 7
)


