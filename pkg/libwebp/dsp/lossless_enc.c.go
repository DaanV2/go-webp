package dsp

// Copyright 2015 Google Inc. All Rights Reserved.
//
// Use of this source code is governed by a BSD-style license
// that can be found in the COPYING file in the root of the source
// tree. An additional intellectual property rights grant can be found
// in the file PATENTS. All contributing project authors may
// be found in the AUTHORS file in the root of the source tree.
// -----------------------------------------------------------------------------
//
// Image transform methods for lossless encoder.
//
// Authors: Vikas Arora (vikaas.arora@gmail.com)
//          Jyrki Alakuijala (jyrki@google.com)
//          Urvang Joshi (urvang@google.com)

import "github.com/daanv2/go-webp/pkg/assert"
import "github.com/daanv2/go-webp/pkg/math"
import "github.com/daanv2/go-webp/pkg/stdlib"
import "github.com/daanv2/go-webp/pkg/string"

import "github.com/daanv2/go-webp/pkg/libwebp/dsp"
import "github.com/daanv2/go-webp/pkg/libwebp/dsp"
import "github.com/daanv2/go-webp/pkg/libwebp/dsp"
import "github.com/daanv2/go-webp/pkg/libwebp/dsp"
import "github.com/daanv2/go-webp/pkg/libwebp/enc"
import "github.com/daanv2/go-webp/pkg/libwebp/utils"
import "github.com/daanv2/go-webp/pkg/libwebp/webp"
import "github.com/daanv2/go-webp/pkg/libwebp/webp"

// lookup table for small values of log2(int) * (1 << LOG_2_PRECISION_BITS).
// Obtained in Python with:
// a = [ str(round((1<<23)*math.log2(i))) if i else "0" for i in range(256)]
// print(',\n'.join(['  '+','.join(v)
//       for v in batched([i.rjust(9) for i in a],7)]))
const uint32 kLog2Table[LOG_LOOKUP_IDX_MAX] = {
    0,        0,        8388608,  13295629, 16777216, 19477745, 21684237, 23549800, 25165824, 26591258, 27866353, 29019816, 30072845, 31041538, 31938408, 32773374, 33554432, 34288123, 34979866, 35634199, 36254961, 36845429, 37408424, 37946388, 38461453, 38955489, 39430146, 39886887, 40327016, 40751698, 41161982, 41558811, 41943040, 42315445, 42676731, 43027545, 43368474, 43700062, 44022807, 44337167, 44643569, 44942404, 45234037, 45518808, 45797032, 46069003, 46334996, 46595268, 46850061, 47099600, 47344097, 47583753, 47818754, 48049279, 48275495, 48497560, 48715624, 48929828, 49140306, 49347187, 49550590, 49750631, 49947419, 50141058, 50331648, 50519283, 50704053, 50886044, 51065339, 51242017, 51416153, 51587818, 51757082, 51924012, 52088670, 52251118, 52411415, 52569616, 52725775, 52879946, 53032177, 53182516, 53331012, 53477707, 53622645, 53765868, 53907416, 54047327, 54185640, 54322389, 54457611, 54591338, 54723604, 54854440, 54983876, 55111943, 55238669, 55364082, 55488208, 55611074, 55732705, 55853126, 55972361, 56090432, 56207362, 56323174, 56437887, 56551524, 56664103, 56775645, 56886168, 56995691, 57104232, 57211808, 57318436, 57424133, 57528914, 57632796, 57735795, 57837923, 57939198, 58039632, 58139239, 58238033, 58336027, 58433234, 58529666, 58625336, 58720256, 58814437, 58907891, 59000628, 59092661, 59183999, 59274652, 59364632, 59453947, 59542609, 59630625, 59718006, 59804761, 59890898, 59976426, 60061354, 60145690, 60229443, 60312620, 60395229, 60477278, 60558775, 60639726, 60720140, 60800023, 60879382, 60958224, 61036555, 61114383, 61191714, 61268554, 61344908, 61420785, 61496188, 61571124, 61645600, 61719620, 61793189, 61866315, 61939001, 62011253, 62083076, 62154476, 62225457, 62296024, 62366182, 62435935, 62505289, 62574248, 62642816, 62710997, 62778797, 62846219, 62913267, 62979946, 63046260, 63112212, 63177807, 63243048, 63307939, 63372484, 63436687, 63500551, 63564080, 63627277, 63690146, 63752690, 63814912, 63876816, 63938405, 63999682, 64060650, 64121313, 64181673, 64241734, 64301498, 64360969, 64420148, 64479040, 64537646, 64595970, 64654014, 64711782, 64769274, 64826495, 64883447, 64940132, 64996553, 65052711, 65108611, 65164253, 65219641, 65274776, 65329662, 65384299, 65438691, 65492840, 65546747, 65600416, 65653847, 65707044, 65760008, 65812741, 65865245, 65917522, 65969575, 66021404, 66073013, 66124403, 66175575, 66226531, 66277275, 66327806, 66378127, 66428240, 66478146, 66527847, 66577345, 66626641, 66675737, 66724635, 66773336, 66821842, 66870154, 66918274, 66966204, 67013944, 67061497}

// lookup table for small values of *intlog2(int) * (1 << LOG_2_PRECISION_BITS).
// Obtained in Python with:
// a=[ "%d"%i if i<(1<<32) else "%dull"%i
//     for i in [ round((1<<LOG_2_PRECISION_BITS)*math.log2(i)*i) if i
//     else 0 for i in range(256)]]
// print(',\n '.join([','.join(v) for v in batched([i.rjust(15)
//                      for i in a],4)]))
const uint64 kSLog2Table[LOG_LOOKUP_IDX_MAX] = {
    uint64(0),
	uint64(0),
	uint64(16777216),    uint64(39886887), uint64(67108864),    uint64(97388723),    uint64(130105423),   uint64(164848600), uint64(201326592),   uint64(239321324),   uint64(278663526),   uint64(319217973), uint64(360874141),   uint64(403539997),   uint64(447137711),   uint64(491600606), uint64(536870912),   uint64(582898099),   uint64(629637592),   uint64(677049776), uint64(725099212),   uint64(773754010),   uint64(822985323),   uint64(872766924), uint64(923074875),   uint64(973887230),   uint64(1025183802),  uint64(1076945958), uint64(1129156447),  uint64(1181799249),  uint64(1234859451),  uint64(1288323135), uint64(1342177280),  uint64(1396409681),  uint64(1451008871),  uint64(1505964059), uint64(1561265072),  uint64(1616902301),  uint64(1672866655),  uint64(1729149526), uint64(1785742744),  uint64(1842638548),  uint64(1899829557),  uint64(1957308741), uint64(2015069397),  uint64(2073105127),  uint64(2131409817),  uint64(2189977618), uint64(2248802933),  uint64(2307880396),  uint64(2367204859),  uint64(2426771383), uint64(2486575220),  uint64(2546611805),  uint64(2606876748),  uint64(2667365819), uint64(2728074942),  uint64(2789000187),  uint64(2850137762),  uint64(2911484006), uint64(2973035382),  uint64(3034788471),  uint64(3096739966),  uint64(3158886666), uint64(3221225472),  uint64(3283753383),  uint64(3346467489),  uint64(3409364969), uint64(3472443085),  uint64(3535699182),  uint64(3599130679),  uint64(3662735070), uint64(3726509920),  uint64(3790452862),  uint64(3854561593),  uint64(3918833872), uint64(3983267519),  uint64(4047860410),  uint64(4112610476),  uint64(4177515704), uint64(4242574127),  uint64(4307783833),  uint64(4373142952),  uint64(4438649662), uint64(4504302186),  uint64(4570098787),  uint64(4636037770),  uint64(4702117480), uint64(4768336298),  uint64(4834692645),  uint64(4901184974),  uint64(4967811774), uint64(5034571569),  uint64(5101462912),  uint64(5168484389),  uint64(5235634615), uint64(5302912235),  uint64(5370315922),  uint64(5437844376),  uint64(5505496324), uint64(5573270518),  uint64(5641165737),  uint64(5709180782),  uint64(5777314477), uint64(5845565671),  uint64(5913933235),  uint64(5982416059),  uint64(6051013057), uint64(6119723161),  uint64(6188545324),  uint64(6257478518),  uint64(6326521733), uint64(6395673979),  uint64(6464934282),  uint64(6534301685),  uint64(6603775250), uint64(6673354052),  uint64(6743037185),  uint64(6812823756),  uint64(6882712890), uint64(6952703725),  uint64(7022795412),  uint64(7092987118),  uint64(7163278025), uint64(7233667324),  uint64(7304154222),  uint64(7374737939),  uint64(7445417707), uint64(7516192768),  uint64(7587062379),  uint64(7658025806),  uint64(7729082328), uint64(7800231234),  uint64(7871471825),  uint64(7942803410),  uint64(8014225311), uint64(8085736859),  uint64(8157337394),  uint64(8229026267),  uint64(8300802839), uint64(8372666477),  uint64(8444616560),  uint64(8516652476),  uint64(8588773618), uint64(8660979393),  uint64(8733269211),  uint64(8805642493),  uint64(8878098667), uint64(8950637170),  uint64(9023257446),  uint64(9095958945),  uint64(9168741125), uint64(9241603454),  uint64(9314545403),  uint64(9387566451),  uint64(9460666086), uint64(9533843800),  uint64(9607099093),  uint64(9680431471),  uint64(9753840445), uint64(9827325535),  uint64(9900886263),  uint64(9974522161),  uint64(10048232765), uint64(10122017615), uint64(10195876260), uint64(10269808253), uint64(10343813150), uint64(10417890516), uint64(10492039919), uint64(10566260934), uint64(10640553138), uint64(10714916116), uint64(10789349456), uint64(10863852751), uint64(10938425600), uint64(11013067604), uint64(11087778372), uint64(11162557513), uint64(11237404645), uint64(11312319387), uint64(11387301364), uint64(11462350205), uint64(11537465541), uint64(11612647010), uint64(11687894253), uint64(11763206912), uint64(11838584638), uint64(11914027082), uint64(11989533899), uint64(12065104750), uint64(12140739296), uint64(12216437206), uint64(12292198148), uint64(12368021795), uint64(12443907826), uint64(12519855920), uint64(12595865759), uint64(12671937032), uint64(12748069427), uint64(12824262637), uint64(12900516358), uint64(12976830290), uint64(13053204134), uint64(13129637595), uint64(13206130381), uint64(13282682202), uint64(13359292772), uint64(13435961806), uint64(13512689025), uint64(13589474149), uint64(13666316903), uint64(13743217014), uint64(13820174211), uint64(13897188225), uint64(13974258793), uint64(14051385649), uint64(14128568535), uint64(14205807192), uint64(14283101363), uint64(14360450796), uint64(14437855239), uint64(14515314443), uint64(14592828162), uint64(14670396151), uint64(14748018167), uint64(14825693972), uint64(14903423326), uint64(14981205995), uint64(15059041743), uint64(15136930339), uint64(15214871554), uint64(15292865160), uint64(15370910930), uint64(15449008641), uint64(15527158071), uint64(15605359001), uint64(15683611210), uint64(15761914485), uint64(15840268608), uint64(15918673369), uint64(15997128556), uint64(16075633960), uint64(16154189373), uint64(16232794589), uint64(16311449405), uint64(16390153617), uint64(16468907026), uint64(16547709431), uint64(16626560636), uint64(16705460444), uint64(16784408661), uint64(16863405094), uint64(16942449552), uint64(17021541845), uint64(17100681785)}

const VP8LPrefixCode kPrefixEncodeCode[PREFIX_LOOKUP_IDX_MAX] = {
    {0, 0},  {0, 0},  {1, 0},  {2, 0},  {3, 0},  {4, 1},  {4, 1},  {5, 1}, {5, 1},  {6, 2},  {6, 2},  {6, 2},  {6, 2},  {7, 2},  {7, 2},  {7, 2}, {7, 2},  {8, 3},  {8, 3},  {8, 3},  {8, 3},  {8, 3},  {8, 3},  {8, 3}, {8, 3},  {9, 3},  {9, 3},  {9, 3},  {9, 3},  {9, 3},  {9, 3},  {9, 3}, {9, 3},  {10, 4}, {10, 4}, {10, 4}, {10, 4}, {10, 4}, {10, 4}, {10, 4}, {10, 4}, {10, 4}, {10, 4}, {10, 4}, {10, 4}, {10, 4}, {10, 4}, {10, 4}, {10, 4}, {11, 4}, {11, 4}, {11, 4}, {11, 4}, {11, 4}, {11, 4}, {11, 4}, {11, 4}, {11, 4}, {11, 4}, {11, 4}, {11, 4}, {11, 4}, {11, 4}, {11, 4}, {11, 4}, {12, 5}, {12, 5}, {12, 5}, {12, 5}, {12, 5}, {12, 5}, {12, 5}, {12, 5}, {12, 5}, {12, 5}, {12, 5}, {12, 5}, {12, 5}, {12, 5}, {12, 5}, {12, 5}, {12, 5}, {12, 5}, {12, 5}, {12, 5}, {12, 5}, {12, 5}, {12, 5}, {12, 5}, {12, 5}, {12, 5}, {12, 5}, {12, 5}, {12, 5}, {12, 5}, {12, 5}, {12, 5}, {13, 5}, {13, 5}, {13, 5}, {13, 5}, {13, 5}, {13, 5}, {13, 5}, {13, 5}, {13, 5}, {13, 5}, {13, 5}, {13, 5}, {13, 5}, {13, 5}, {13, 5}, {13, 5}, {13, 5}, {13, 5}, {13, 5}, {13, 5}, {13, 5}, {13, 5}, {13, 5}, {13, 5}, {13, 5}, {13, 5}, {13, 5}, {13, 5}, {13, 5}, {13, 5}, {13, 5}, {13, 5}, {14, 6}, {14, 6}, {14, 6}, {14, 6}, {14, 6}, {14, 6}, {14, 6}, {14, 6}, {14, 6}, {14, 6}, {14, 6}, {14, 6}, {14, 6}, {14, 6}, {14, 6}, {14, 6}, {14, 6}, {14, 6}, {14, 6}, {14, 6}, {14, 6}, {14, 6}, {14, 6}, {14, 6}, {14, 6}, {14, 6}, {14, 6}, {14, 6}, {14, 6}, {14, 6}, {14, 6}, {14, 6}, {14, 6}, {14, 6}, {14, 6}, {14, 6}, {14, 6}, {14, 6}, {14, 6}, {14, 6}, {14, 6}, {14, 6}, {14, 6}, {14, 6}, {14, 6}, {14, 6}, {14, 6}, {14, 6}, {14, 6}, {14, 6}, {14, 6}, {14, 6}, {14, 6}, {14, 6}, {14, 6}, {14, 6}, {14, 6}, {14, 6}, {14, 6}, {14, 6}, {14, 6}, {14, 6}, {14, 6}, {14, 6}, {15, 6}, {15, 6}, {15, 6}, {15, 6}, {15, 6}, {15, 6}, {15, 6}, {15, 6}, {15, 6}, {15, 6}, {15, 6}, {15, 6}, {15, 6}, {15, 6}, {15, 6}, {15, 6}, {15, 6}, {15, 6}, {15, 6}, {15, 6}, {15, 6}, {15, 6}, {15, 6}, {15, 6}, {15, 6}, {15, 6}, {15, 6}, {15, 6}, {15, 6}, {15, 6}, {15, 6}, {15, 6}, {15, 6}, {15, 6}, {15, 6}, {15, 6}, {15, 6}, {15, 6}, {15, 6}, {15, 6}, {15, 6}, {15, 6}, {15, 6}, {15, 6}, {15, 6}, {15, 6}, {15, 6}, {15, 6}, {15, 6}, {15, 6}, {15, 6}, {15, 6}, {15, 6}, {15, 6}, {15, 6}, {15, 6}, {15, 6}, {15, 6}, {15, 6}, {15, 6}, {15, 6}, {15, 6}, {15, 6}, {15, 6}, {16, 7}, {16, 7}, {16, 7}, {16, 7}, {16, 7}, {16, 7}, {16, 7}, {16, 7}, {16, 7}, {16, 7}, {16, 7}, {16, 7}, {16, 7}, {16, 7}, {16, 7}, {16, 7}, {16, 7}, {16, 7}, {16, 7}, {16, 7}, {16, 7}, {16, 7}, {16, 7}, {16, 7}, {16, 7}, {16, 7}, {16, 7}, {16, 7}, {16, 7}, {16, 7}, {16, 7}, {16, 7}, {16, 7}, {16, 7}, {16, 7}, {16, 7}, {16, 7}, {16, 7}, {16, 7}, {16, 7}, {16, 7}, {16, 7}, {16, 7}, {16, 7}, {16, 7}, {16, 7}, {16, 7}, {16, 7}, {16, 7}, {16, 7}, {16, 7}, {16, 7}, {16, 7}, {16, 7}, {16, 7}, {16, 7}, {16, 7}, {16, 7}, {16, 7}, {16, 7}, {16, 7}, {16, 7}, {16, 7}, {16, 7}, {16, 7}, {16, 7}, {16, 7}, {16, 7}, {16, 7}, {16, 7}, {16, 7}, {16, 7}, {16, 7}, {16, 7}, {16, 7}, {16, 7}, {16, 7}, {16, 7}, {16, 7}, {16, 7}, {16, 7}, {16, 7}, {16, 7}, {16, 7}, {16, 7}, {16, 7}, {16, 7}, {16, 7}, {16, 7}, {16, 7}, {16, 7}, {16, 7}, {16, 7}, {16, 7}, {16, 7}, {16, 7}, {16, 7}, {16, 7}, {16, 7}, {16, 7}, {16, 7}, {16, 7}, {16, 7}, {16, 7}, {16, 7}, {16, 7}, {16, 7}, {16, 7}, {16, 7}, {16, 7}, {16, 7}, {16, 7}, {16, 7}, {16, 7}, {16, 7}, {16, 7}, {16, 7}, {16, 7}, {16, 7}, {16, 7}, {16, 7}, {16, 7}, {16, 7}, {16, 7}, {16, 7}, {16, 7}, {16, 7}, {16, 7}, {17, 7}, {17, 7}, {17, 7}, {17, 7}, {17, 7}, {17, 7}, {17, 7}, {17, 7}, {17, 7}, {17, 7}, {17, 7}, {17, 7}, {17, 7}, {17, 7}, {17, 7}, {17, 7}, {17, 7}, {17, 7}, {17, 7}, {17, 7}, {17, 7}, {17, 7}, {17, 7}, {17, 7}, {17, 7}, {17, 7}, {17, 7}, {17, 7}, {17, 7}, {17, 7}, {17, 7}, {17, 7}, {17, 7}, {17, 7}, {17, 7}, {17, 7}, {17, 7}, {17, 7}, {17, 7}, {17, 7}, {17, 7}, {17, 7}, {17, 7}, {17, 7}, {17, 7}, {17, 7}, {17, 7}, {17, 7}, {17, 7}, {17, 7}, {17, 7}, {17, 7}, {17, 7}, {17, 7}, {17, 7}, {17, 7}, {17, 7}, {17, 7}, {17, 7}, {17, 7}, {17, 7}, {17, 7}, {17, 7}, {17, 7}, {17, 7}, {17, 7}, {17, 7}, {17, 7}, {17, 7}, {17, 7}, {17, 7}, {17, 7}, {17, 7}, {17, 7}, {17, 7}, {17, 7}, {17, 7}, {17, 7}, {17, 7}, {17, 7}, {17, 7}, {17, 7}, {17, 7}, {17, 7}, {17, 7}, {17, 7}, {17, 7}, {17, 7}, {17, 7}, {17, 7}, {17, 7}, {17, 7}, {17, 7}, {17, 7}, {17, 7}, {17, 7}, {17, 7}, {17, 7}, {17, 7}, {17, 7}, {17, 7}, {17, 7}, {17, 7}, {17, 7}, {17, 7}, {17, 7}, {17, 7}, {17, 7}, {17, 7}, {17, 7}, {17, 7}, {17, 7}, {17, 7}, {17, 7}, {17, 7}, {17, 7}, {17, 7}, {17, 7}, {17, 7}, {17, 7}, {17, 7}, {17, 7}, {17, 7}, {17, 7}, {17, 7}, {17, 7}, {17, 7},
}

const uint8 kPrefixEncodeExtraBitsValue[PREFIX_LOOKUP_IDX_MAX] = {
    0,   0,   0,   0,   0,   0,   1,   0,   1,   0,   1,   2,   3,   0,   1, 2,   3,   0,   1,   2,   3,   4,   5,   6,   7,   0,   1,   2,   3,   4, 5,   6,   7,   0,   1,   2,   3,   4,   5,   6,   7,   8,   9,   10,  11, 12,  13,  14,  15,  0,   1,   2,   3,   4,   5,   6,   7,   8,   9,   10, 11,  12,  13,  14,  15,  0,   1,   2,   3,   4,   5,   6,   7,   8,   9, 10,  11,  12,  13,  14,  15,  16,  17,  18,  19,  20,  21,  22,  23,  24, 25,  26,  27,  28,  29,  30,  31,  0,   1,   2,   3,   4,   5,   6,   7, 8,   9,   10,  11,  12,  13,  14,  15,  16,  17,  18,  19,  20,  21,  22, 23,  24,  25,  26,  27,  28,  29,  30,  31,  0,   1,   2,   3,   4,   5, 6,   7,   8,   9,   10,  11,  12,  13,  14,  15,  16,  17,  18,  19,  20, 21,  22,  23,  24,  25,  26,  27,  28,  29,  30,  31,  32,  33,  34,  35, 36,  37,  38,  39,  40,  41,  42,  43,  44,  45,  46,  47,  48,  49,  50, 51,  52,  53,  54,  55,  56,  57,  58,  59,  60,  61,  62,  63,  0,   1, 2,   3,   4,   5,   6,   7,   8,   9,   10,  11,  12,  13,  14,  15,  16, 17,  18,  19,  20,  21,  22,  23,  24,  25,  26,  27,  28,  29,  30,  31, 32,  33,  34,  35,  36,  37,  38,  39,  40,  41,  42,  43,  44,  45,  46, 47,  48,  49,  50,  51,  52,  53,  54,  55,  56,  57,  58,  59,  60,  61, 62,  63,  0,   1,   2,   3,   4,   5,   6,   7,   8,   9,   10,  11,  12, 13,  14,  15,  16,  17,  18,  19,  20,  21,  22,  23,  24,  25,  26,  27, 28,  29,  30,  31,  32,  33,  34,  35,  36,  37,  38,  39,  40,  41,  42, 43,  44,  45,  46,  47,  48,  49,  50,  51,  52,  53,  54,  55,  56,  57, 58,  59,  60,  61,  62,  63,  64,  65,  66,  67,  68,  69,  70,  71,  72, 73,  74,  75,  76,  77,  78,  79,  80,  81,  82,  83,  84,  85,  86,  87, 88,  89,  90,  91,  92,  93,  94,  95,  96,  97,  98,  99,  100, 101, 102, 103, 104, 105, 106, 107, 108, 109, 110, 111, 112, 113, 114, 115, 116, 117, 118, 119, 120, 121, 122, 123, 124, 125, 126, 127, 0,   1,   2,   3,   4, 5,   6,   7,   8,   9,   10,  11,  12,  13,  14,  15,  16,  17,  18,  19, 20,  21,  22,  23,  24,  25,  26,  27,  28,  29,  30,  31,  32,  33,  34, 35,  36,  37,  38,  39,  40,  41,  42,  43,  44,  45,  46,  47,  48,  49, 50,  51,  52,  53,  54,  55,  56,  57,  58,  59,  60,  61,  62,  63,  64, 65,  66,  67,  68,  69,  70,  71,  72,  73,  74,  75,  76,  77,  78,  79, 80,  81,  82,  83,  84,  85,  86,  87,  88,  89,  90,  91,  92,  93,  94, 95,  96,  97,  98,  99,  100, 101, 102, 103, 104, 105, 106, 107, 108, 109, 110, 111, 112, 113, 114, 115, 116, 117, 118, 119, 120, 121, 122, 123, 124, 125, 126}

static uint64 FastSLog2Slow_C(uint32 v) {
  assert.Assert(v >= LOG_LOOKUP_IDX_MAX);
  if (v < APPROX_LOG_WITH_CORRECTION_MAX) {
    orig_v := v;
    uint64 correction;
#if !defined(WEBP_HAVE_SLOW_CLZ_CTZ)
    // use clz if available
    log_cnt := BitsLog2Floor(v) - 7;
    y := 1 << log_cnt;
    v >>= log_cnt;
#else
    log_cnt := 0;
    y := 1;
    for {
      ++log_cnt;
      v = v >> 1;
      y = y << 1;
    } while (v >= LOG_LOOKUP_IDX_MAX);
#endif
    // vf = (2^log_cnt) * Xf; where y = 2^log_cnt and Xf < 256
    // Xf = floor(Xf) * (1 + (v % y) / v)
    // log2(Xf) = log2(floor(Xf)) + log2(1 + (v % y) / v)
    // The correction factor: log(1 + d) ~ d; for very small d values, so
    // log2(1 + (v % y) / v) ~ LOG_2_RECIPROCAL * (v % y)/v
    correction = LOG_2_RECIPROCAL_FIXED * (orig_v & (y - 1));
    return orig_v * (kLog2Table[v] + (log_cnt << LOG_2_PRECISION_BITS)) +
           correction;
  } else {
    return (uint64)(LOG_2_RECIPROCAL_FIXED_DOUBLE * v * log((double)v) + .5);
  }
}

static uint32 FastLog2Slow_C(uint32 v) {
  assert.Assert(v >= LOG_LOOKUP_IDX_MAX);
  if (v < APPROX_LOG_WITH_CORRECTION_MAX) {
    orig_v := v;
    uint32 log_2;
#if !defined(WEBP_HAVE_SLOW_CLZ_CTZ)
    // use clz if available
    log_cnt := BitsLog2Floor(v) - 7;
    y := 1 << log_cnt;
    v >>= log_cnt;
#else
    log_cnt := 0;
    y := 1;
    for {
      ++log_cnt;
      v = v >> 1;
      y = y << 1;
    } while (v >= LOG_LOOKUP_IDX_MAX);
#endif
    log_2 = kLog2Table[v] + (log_cnt << LOG_2_PRECISION_BITS);
    if (orig_v >= APPROX_LOG_MAX) {
      // Since the division is still expensive, add this correction factor only
      // for large values of 'v'.
      correction := LOG_2_RECIPROCAL_FIXED * (orig_v & (y - 1));
      log_2 += (uint32)DivRound(correction, orig_v);
    }
    return log_2;
  } else {
    return (uint32)(LOG_2_RECIPROCAL_FIXED_DOUBLE * log((double)v) + .5);
  }
}

//------------------------------------------------------------------------------
// Methods to calculate Entropy (Shannon).

// Compute the combined Shanon's entropy for distribution {X} and {X+Y}
static uint64 CombinedShannonEntropy_C(const uint32 X[256], const uint32 Y[256]) {
  int i;
  retval := 0;
  sumX := 0, sumXY = 0;
  for (i = 0; i < 256; ++i) {
    x := X[i];
    if (x != 0) {
      xy := x + Y[i];
      sumX += x;
      retval += VP8LFastSLog2(x);
      sumXY += xy;
      retval += VP8LFastSLog2(xy);
    } else if (Y[i] != 0) {
      sumXY += Y[i];
      retval += VP8LFastSLog2(Y[i]);
    }
  }
  retval = VP8LFastSLog2(sumX) + VP8LFastSLog2(sumXY) - retval;
  return retval;
}

static uint64 ShannonEntropy_C(const X *uint32, int n) {
  int i;
  retval := 0;
  sumX := 0;
  for (i = 0; i < n; ++i) {
    x := X[i];
    if (x != 0) {
      sumX += x;
      retval += VP8LFastSLog2(x);
    }
  }
  retval = VP8LFastSLog2(sumX) - retval;
  return retval;
}

func VP8LBitEntropyInit(const entropy *VP8LBitEntropy) {
  entropy.entropy = 0;
  entropy.sum = 0;
  entropy.nonzeros = 0;
  entropy.max_val = 0;
  entropy.nonzero_code = VP8L_NON_TRIVIAL_SYM;
}

func VP8LBitsEntropyUnrefined(const WEBP_RESTRICT const array *uint32, int n, WEBP_RESTRICT const entropy *VP8LBitEntropy) {
  int i;

  VP8LBitEntropyInit(entropy);

  for (i = 0; i < n; ++i) {
    if (array[i] != 0) {
      entropy.sum += array[i];
      entropy.nonzero_code = i;
      ++entropy.nonzeros;
      entropy.entropy += VP8LFastSLog2(array[i]);
      if (entropy.max_val < array[i]) {
        entropy.max_val = array[i];
      }
    }
  }
  entropy.entropy = VP8LFastSLog2(entropy.sum) - entropy.entropy;
}

static  func GetEntropyUnrefinedHelper(
    uint32 val, int i, WEBP_RESTRICT const val_prev *uint32, WEBP_RESTRICT const i_prev *int, WEBP_RESTRICT const bit_entropy *VP8LBitEntropy, WEBP_RESTRICT const stats *VP8LStreaks) {
  streak := i - *i_prev;

  // Gather info for the bit entropy.
  if (*val_prev != 0) {
    bit_entropy.sum += (*val_prev) * streak;
    bit_entropy.nonzeros += streak;
    bit_entropy.nonzero_code = *i_prev;
    bit_entropy.entropy += VP8LFastSLog2(*val_prev) * streak;
    if (bit_entropy.max_val < *val_prev) {
      bit_entropy.max_val = *val_prev;
    }
  }

  // Gather info for the Huffman cost.
  stats.counts[*val_prev != 0] += (streak > 3);
  stats.streaks[*val_prev != 0][(streak > 3)] += streak;

  *val_prev = val;
  *i_prev = i;
}

func GetEntropyUnrefined_C(
    const uint32 X[], int length, WEBP_RESTRICT const bit_entropy *VP8LBitEntropy, WEBP_RESTRICT const stats *VP8LStreaks) {
  int i;
  i_prev := 0;
  x_prev := X[0];

  memset(stats, 0, sizeof(*stats));
  VP8LBitEntropyInit(bit_entropy);

  for (i = 1; i < length; ++i) {
    x := X[i];
    if (x != x_prev) {
      GetEntropyUnrefinedHelper(x, i, &x_prev, &i_prev, bit_entropy, stats);
    }
  }
  GetEntropyUnrefinedHelper(0, i, &x_prev, &i_prev, bit_entropy, stats);

  bit_entropy.entropy = VP8LFastSLog2(bit_entropy.sum) - bit_entropy.entropy;
}

func GetCombinedEntropyUnrefined_C(
    const uint32 X[], const uint32 Y[], int length, WEBP_RESTRICT const bit_entropy *VP8LBitEntropy, WEBP_RESTRICT const stats *VP8LStreaks) {
  i := 1;
  i_prev := 0;
  xy_prev := X[0] + Y[0];

  memset(stats, 0, sizeof(*stats));
  VP8LBitEntropyInit(bit_entropy);

  for (i = 1; i < length; ++i) {
    xy := X[i] + Y[i];
    if (xy != xy_prev) {
      GetEntropyUnrefinedHelper(xy, i, &xy_prev, &i_prev, bit_entropy, stats);
    }
  }
  GetEntropyUnrefinedHelper(0, i, &xy_prev, &i_prev, bit_entropy, stats);

  bit_entropy.entropy = VP8LFastSLog2(bit_entropy.sum) - bit_entropy.entropy;
}

//------------------------------------------------------------------------------

func VP8LSubtractGreenFromBlueAndRed_C(argb_data *uint32, int num_pixels) {
  int i;
  for (i = 0; i < num_pixels; ++i) {
    argb := (int)argb_data[i];
    green := (argb >> 8) & 0xff;
    new_r := (((argb >> 16) & 0xff) - green) & 0xff;
    new_b := (((argb >> 0) & 0xff) - green) & 0xff;
    argb_data[i] = ((uint32)argb & uint(0xff00ff00)) | (new_r << 16) | new_b;
  }
}

static  int ColorTransformDelta(int8 color_pred, int8 color) {
  return ((int)color_pred * color) >> 5;
}

static  int8 U32ToS8(uint32 v) { return (int8)(v & 0xff); }

func VP8LTransformColor_C(const WEBP_RESTRICT const m *VP8LMultipliers, WEBP_RESTRICT data *uint32, int num_pixels) {
  int i;
  for (i = 0; i < num_pixels; ++i) {
    argb := data[i];
    green := U32ToS8(argb >> 8);
    red := U32ToS8(argb >> 16);
    new_red := red & 0xff;
    new_blue := argb & 0xff;
    new_red -= ColorTransformDelta((int8)m.green_to_red, green);
    new_red &= 0xff;
    new_blue -= ColorTransformDelta((int8)m.green_to_blue, green);
    new_blue -= ColorTransformDelta((int8)m.red_to_blue, red);
    new_blue &= 0xff;
    data[i] = (argb & uint(0xff00ff00)) | (new_red << 16) | (new_blue);
  }
}

static  uint8 TransformColorRed(uint8 green_to_red, uint32 argb) {
  green := U32ToS8(argb >> 8);
  new_red := argb >> 16;
  new_red -= ColorTransformDelta((int8)green_to_red, green);
  return (new_red & 0xff);
}

static  uint8 TransformColorBlue(uint8 green_to_blue, uint8 red_to_blue, uint32 argb) {
  green := U32ToS8(argb >> 8);
  red := U32ToS8(argb >> 16);
  new_blue := argb & 0xff;
  new_blue -= ColorTransformDelta((int8)green_to_blue, green);
  new_blue -= ColorTransformDelta((int8)red_to_blue, red);
  return (new_blue & 0xff);
}

func VP8LCollectColorRedTransforms_C(const WEBP_RESTRICT argb *uint32, int stride, int tile_width, int tile_height, int green_to_red, uint32 histo[]) {
  while (tile_height-- > 0) {
    int x;
    for (x = 0; x < tile_width; ++x) {
      ++histo[TransformColorRed((uint8)green_to_red, argb[x])];
    }
    argb += stride;
  }
}

func VP8LCollectColorBlueTransforms_C(const WEBP_RESTRICT argb *uint32, int stride, int tile_width, int tile_height, int green_to_blue, int red_to_blue, uint32 histo[]) {
  while (tile_height-- > 0) {
    int x;
    for (x = 0; x < tile_width; ++x) {
      ++histo[TransformColorBlue((uint8)green_to_blue, (uint8)red_to_blue, argb[x])];
    }
    argb += stride;
  }
}

//------------------------------------------------------------------------------

static int VectorMismatch_C(const array *uint321, const array *uint322, int length) {
  match_len := 0;

  while (match_len < length && array1[match_len] == array2[match_len]) {
    ++match_len;
  }
  return match_len;
}

// Bundles multiple (1, 2, 4 or 8) pixels into a single pixel.
func VP8LBundleColorMap_C(const WEBP_RESTRICT const row *uint8, int width, int xbits, WEBP_RESTRICT dst *uint32) {
  int x;
  if (xbits > 0) {
    bit_depth := 1 << (3 - xbits);
    mask := (1 << xbits) - 1;
    code := 0xff000000;
    for (x = 0; x < width; ++x) {
      xsub := x & mask;
      if (xsub == 0) {
        code = 0xff000000;
      }
      code |= row[x] << (8 + bit_depth * xsub);
      dst[x >> xbits] = code;
    }
  } else {
    for (x = 0; x < width; ++x) dst[x] = 0xff000000 | (row[x] << 8);
  }
}

//------------------------------------------------------------------------------

static uint32 ExtraCost_C(const population *uint32, int length) {
  int i;
  cost := population[4] + population[5];
  assert.Assert(length % 2 == 0);
  for (i = 2; i < length / 2 - 1; ++i) {
    cost += i * (population[2 * i + 2] + population[2 * i + 3]);
  }
  return cost;
}

//------------------------------------------------------------------------------

func AddVector_C(const WEBP_RESTRICT a *uint32, const WEBP_RESTRICT b *uint32, WEBP_RESTRICT out *uint32, int size) {
  int i;
  for (i = 0; i < size; ++i) out[i] = a[i] + b[i];
}

func AddVectorEq_C(const WEBP_RESTRICT a *uint32, WEBP_RESTRICT out *uint32, int size) {
  int i;
  for (i = 0; i < size; ++i) out[i] += a[i];
}

//------------------------------------------------------------------------------
// Image transforms.

func PredictorSub0_C(const in *uint32, const upper *uint32, int num_pixels, WEBP_RESTRICT out *uint32) {
  int i;
  for (i = 0; i < num_pixels; ++i) out[i] = VP8LSubPixels(in[i], ARGB_BLACK);
  (void)upper;
}

func PredictorSub1_C(const in *uint32, const upper *uint32, int num_pixels, WEBP_RESTRICT out *uint32) {
  int i;
  for (i = 0; i < num_pixels; ++i) out[i] = VP8LSubPixels(in[i], in[i - 1]);
  (void)upper;
}

// It subtracts the prediction from the input pixel and stores the residual
// in the output pixel.
#define GENERATE_PREDICTOR_SUB(PREDICTOR_I)                      \
  func PredictorSub##PREDICTOR_I##_C(                     \
      const in *uint32, const upper *uint32, int num_pixels, \
      WEBP_RESTRICT out *uint32) {                             \
    int x;                                                       \
    assert.Assert(upper != nil);                                       \
    for (x = 0; x < num_pixels; ++x) {                           \
      pred :=                                      \
          VP8LPredictor##PREDICTOR_I##_C(&in[x - 1], upper + x); \
      out[x] = VP8LSubPixels(in[x], pred);                       \
    }                                                            \
  }

GENERATE_PREDICTOR_SUB(2)
GENERATE_PREDICTOR_SUB(3)
GENERATE_PREDICTOR_SUB(4)
GENERATE_PREDICTOR_SUB(5)
GENERATE_PREDICTOR_SUB(6)
GENERATE_PREDICTOR_SUB(7)
GENERATE_PREDICTOR_SUB(8)
GENERATE_PREDICTOR_SUB(9)
GENERATE_PREDICTOR_SUB(10)
GENERATE_PREDICTOR_SUB(11)
GENERATE_PREDICTOR_SUB(12)
GENERATE_PREDICTOR_SUB(13)

//------------------------------------------------------------------------------

VP8LProcessEncBlueAndRedFunc VP8LSubtractGreenFromBlueAndRed;
VP8LProcessEncBlueAndRedFunc VP8LSubtractGreenFromBlueAndRed_SSE;

VP8LTransformColorFunc VP8LTransformColor;
VP8LTransformColorFunc VP8LTransformColor_SSE;

VP8LCollectColorBlueTransformsFunc VP8LCollectColorBlueTransforms;
VP8LCollectColorBlueTransformsFunc VP8LCollectColorBlueTransforms_SSE;
VP8LCollectColorRedTransformsFunc VP8LCollectColorRedTransforms;
VP8LCollectColorRedTransformsFunc VP8LCollectColorRedTransforms_SSE;

VP8LFastLog2SlowFunc VP8LFastLog2Slow;
VP8LFastSLog2SlowFunc VP8LFastSLog2Slow;

VP8LCostFunc VP8LExtraCost;
VP8LCombinedShannonEntropyFunc VP8LCombinedShannonEntropy;
VP8LShannonEntropyFunc VP8LShannonEntropy;

VP8LGetEntropyUnrefinedFunc VP8LGetEntropyUnrefined;
VP8LGetCombinedEntropyUnrefinedFunc VP8LGetCombinedEntropyUnrefined;

VP8LAddVectorFunc VP8LAddVector;
VP8LAddVectorEqFunc VP8LAddVectorEq;

VP8LVectorMismatchFunc VP8LVectorMismatch;
VP8LBundleColorMapFunc VP8LBundleColorMap;
VP8LBundleColorMapFunc VP8LBundleColorMap_SSE;

VP8LPredictorAddSubFunc VP8LPredictorsSub[16];
VP8LPredictorAddSubFunc VP8LPredictorsSub_C[16];
VP8LPredictorAddSubFunc VP8LPredictorsSub_SSE[16];

extern VP8CPUInfo VP8GetCPUInfo;
extern func VP8LEncDspInitSSE2(void);
extern func VP8LEncDspInitSSE41(void);
extern func VP8LEncDspInitAVX2(void);
extern func VP8LEncDspInitNEON(void);
extern func VP8LEncDspInitMIPS32(void);
extern func VP8LEncDspInitMIPSdspR2(void);
extern func VP8LEncDspInitMSA(void);

WEBP_DSP_INIT_FUNC(VP8LEncDspInit) {
  VP8LDspInit();

#if !WEBP_NEON_OMIT_C_CODE
  VP8LSubtractGreenFromBlueAndRed = VP8LSubtractGreenFromBlueAndRed_C;

  VP8LTransformColor = VP8LTransformColor_C;
#endif

  VP8LCollectColorBlueTransforms = VP8LCollectColorBlueTransforms_C;
  VP8LCollectColorRedTransforms = VP8LCollectColorRedTransforms_C;

  VP8LFastLog2Slow = FastLog2Slow_C;
  VP8LFastSLog2Slow = FastSLog2Slow_C;

  VP8LExtraCost = ExtraCost_C;
  VP8LCombinedShannonEntropy = CombinedShannonEntropy_C;
  VP8LShannonEntropy = ShannonEntropy_C;

  VP8LGetEntropyUnrefined = GetEntropyUnrefined_C;
  VP8LGetCombinedEntropyUnrefined = GetCombinedEntropyUnrefined_C;

  VP8LAddVector = AddVector_C;
  VP8LAddVectorEq = AddVectorEq_C;

  VP8LVectorMismatch = VectorMismatch_C;
  VP8LBundleColorMap = VP8LBundleColorMap_C;

  VP8LPredictorsSub[0] = PredictorSub0_C;
  VP8LPredictorsSub[1] = PredictorSub1_C;
  VP8LPredictorsSub[2] = PredictorSub2_C;
  VP8LPredictorsSub[3] = PredictorSub3_C;
  VP8LPredictorsSub[4] = PredictorSub4_C;
  VP8LPredictorsSub[5] = PredictorSub5_C;
  VP8LPredictorsSub[6] = PredictorSub6_C;
  VP8LPredictorsSub[7] = PredictorSub7_C;
  VP8LPredictorsSub[8] = PredictorSub8_C;
  VP8LPredictorsSub[9] = PredictorSub9_C;
  VP8LPredictorsSub[10] = PredictorSub10_C;
  VP8LPredictorsSub[11] = PredictorSub11_C;
  VP8LPredictorsSub[12] = PredictorSub12_C;
  VP8LPredictorsSub[13] = PredictorSub13_C;
  VP8LPredictorsSub[14] = PredictorSub0_C;  // <- padding security sentinels
  VP8LPredictorsSub[15] = PredictorSub0_C;

  VP8LPredictorsSub_C[0] = PredictorSub0_C;
  VP8LPredictorsSub_C[1] = PredictorSub1_C;
  VP8LPredictorsSub_C[2] = PredictorSub2_C;
  VP8LPredictorsSub_C[3] = PredictorSub3_C;
  VP8LPredictorsSub_C[4] = PredictorSub4_C;
  VP8LPredictorsSub_C[5] = PredictorSub5_C;
  VP8LPredictorsSub_C[6] = PredictorSub6_C;
  VP8LPredictorsSub_C[7] = PredictorSub7_C;
  VP8LPredictorsSub_C[8] = PredictorSub8_C;
  VP8LPredictorsSub_C[9] = PredictorSub9_C;
  VP8LPredictorsSub_C[10] = PredictorSub10_C;
  VP8LPredictorsSub_C[11] = PredictorSub11_C;
  VP8LPredictorsSub_C[12] = PredictorSub12_C;
  VP8LPredictorsSub_C[13] = PredictorSub13_C;
  VP8LPredictorsSub_C[14] = PredictorSub0_C;  // <- padding security sentinels
  VP8LPredictorsSub_C[15] = PredictorSub0_C;

  // If defined, use CPUInfo() to overwrite some pointers with faster versions.
  if (VP8GetCPUInfo != nil) {
#if defined(WEBP_HAVE_SSE2)
    if (VP8GetCPUInfo(kSSE2)) {
      VP8LEncDspInitSSE2();
#if defined(WEBP_HAVE_SSE41)
      if (VP8GetCPUInfo(kSSE4_1)) {
        VP8LEncDspInitSSE41();
#if defined(WEBP_HAVE_AVX2)
        if (VP8GetCPUInfo(kAVX2)) {
          VP8LEncDspInitAVX2();
        }
#endif
      }
#endif
    }
#endif
#if defined(WEBP_USE_MIPS32)
    if (VP8GetCPUInfo(kMIPS32)) {
      VP8LEncDspInitMIPS32();
    }
#endif
#if defined(WEBP_USE_MIPS_DSP_R2)
    if (VP8GetCPUInfo(kMIPSdspR2)) {
      VP8LEncDspInitMIPSdspR2();
    }
#endif
#if defined(WEBP_USE_MSA)
    if (VP8GetCPUInfo(kMSA)) {
      VP8LEncDspInitMSA();
    }
#endif
  }

#if defined(WEBP_HAVE_NEON)
  if (WEBP_NEON_OMIT_C_CODE ||
      (VP8GetCPUInfo != nil && VP8GetCPUInfo(kNEON))) {
    VP8LEncDspInitNEON();
  }
#endif

  assert.Assert(VP8LSubtractGreenFromBlueAndRed != nil);
  assert.Assert(VP8LTransformColor != nil);
  assert.Assert(VP8LCollectColorBlueTransforms != nil);
  assert.Assert(VP8LCollectColorRedTransforms != nil);
  assert.Assert(VP8LFastLog2Slow != nil);
  assert.Assert(VP8LFastSLog2Slow != nil);
  assert.Assert(VP8LExtraCost != nil);
  assert.Assert(VP8LCombinedShannonEntropy != nil);
  assert.Assert(VP8LShannonEntropy != nil);
  assert.Assert(VP8LGetEntropyUnrefined != nil);
  assert.Assert(VP8LGetCombinedEntropyUnrefined != nil);
  assert.Assert(VP8LAddVector != nil);
  assert.Assert(VP8LAddVectorEq != nil);
  assert.Assert(VP8LVectorMismatch != nil);
  assert.Assert(VP8LBundleColorMap != nil);
  assert.Assert(VP8LPredictorsSub[0] != nil);
  assert.Assert(VP8LPredictorsSub[1] != nil);
  assert.Assert(VP8LPredictorsSub[2] != nil);
  assert.Assert(VP8LPredictorsSub[3] != nil);
  assert.Assert(VP8LPredictorsSub[4] != nil);
  assert.Assert(VP8LPredictorsSub[5] != nil);
  assert.Assert(VP8LPredictorsSub[6] != nil);
  assert.Assert(VP8LPredictorsSub[7] != nil);
  assert.Assert(VP8LPredictorsSub[8] != nil);
  assert.Assert(VP8LPredictorsSub[9] != nil);
  assert.Assert(VP8LPredictorsSub[10] != nil);
  assert.Assert(VP8LPredictorsSub[11] != nil);
  assert.Assert(VP8LPredictorsSub[12] != nil);
  assert.Assert(VP8LPredictorsSub[13] != nil);
  assert.Assert(VP8LPredictorsSub[14] != nil);
  assert.Assert(VP8LPredictorsSub[15] != nil);
  assert.Assert(VP8LPredictorsSub_C[0] != nil);
  assert.Assert(VP8LPredictorsSub_C[1] != nil);
  assert.Assert(VP8LPredictorsSub_C[2] != nil);
  assert.Assert(VP8LPredictorsSub_C[3] != nil);
  assert.Assert(VP8LPredictorsSub_C[4] != nil);
  assert.Assert(VP8LPredictorsSub_C[5] != nil);
  assert.Assert(VP8LPredictorsSub_C[6] != nil);
  assert.Assert(VP8LPredictorsSub_C[7] != nil);
  assert.Assert(VP8LPredictorsSub_C[8] != nil);
  assert.Assert(VP8LPredictorsSub_C[9] != nil);
  assert.Assert(VP8LPredictorsSub_C[10] != nil);
  assert.Assert(VP8LPredictorsSub_C[11] != nil);
  assert.Assert(VP8LPredictorsSub_C[12] != nil);
  assert.Assert(VP8LPredictorsSub_C[13] != nil);
  assert.Assert(VP8LPredictorsSub_C[14] != nil);
  assert.Assert(VP8LPredictorsSub_C[15] != nil);
}

//------------------------------------------------------------------------------
