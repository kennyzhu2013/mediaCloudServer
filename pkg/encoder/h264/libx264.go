// Implements cgo bindings for [x264](https://www.videolan.org/developers/x264.html) library.
package h264

/*
#cgo pkg-config: x264
#cgo CFLAGS: -Wall -O3

#include "stdint.h"
#include "x264.h"
#include <stdlib.h>
*/
import "C"
import "unsafe"

const Build = C.X264_BUILD

/* T is opaque handler for encoder */
type T struct{}

/****************************************************************************
 * NAL structure and functions
 ****************************************************************************/

/* enum nal_unit_type_e */
const (
	NalUnknown  = 0
	NalSlice    = 1
	NalSliceDpa = 2
	NalSliceDpb = 3
	NalSliceDpc = 4
	NalSliceIdr = 5 /* ref_idc != 0 */
	NalSei      = 6 /* ref_idc == 0 */
	NalSps      = 7
	NalPps      = 8
	NalAud      = 9
	NalFiller   = 12
	/* ref_idc == 0 for 6,9,10,11,12 */
)

/* enum nal_priority_e */
const (
	NalPriorityDisposable = 0
	NalPriorityLow        = 1
	NalPriorityHigh       = 2
	NalPriorityHighest    = 3
)

/* The data within the payload is already NAL-encapsulated; the ref_idc and type
 * are merely in the struct for easy access by the calling application.
 * All data returned in an x264_nal_t, including the data in p_payload, is no longer
 * valid after the next call to x264_encoder_encode.  Thus it must be used or copied
 * before calling x264_encoder_encode or x264_encoder_headers again. */
type Nal struct {
	IRefIdc        int32 /* nal_priority_e */
	IType          int32 /* nal_unit_type_e */
	BLongStartcode int32
	IFirstMb       int32 /* If this NAL is a slice, the index of the first MB in the slice. */
	ILastMb        int32 /* If this NAL is a slice, the index of the last MB in the slice. */

	/* Size of payload (including any padding) in bytes. */
	IPayload int32
	/* If param->b_annexb is set, Annex-B bytestream with startcode.
	 * Otherwise, startcode is replaced with a 4-byte size.
	 * This size is the size used in mp4/similar muxing; it is equal to i_payload-4 */
	/* C.uint8_t */
	PPayload unsafe.Pointer

	/* Size of padding in bytes. */
	IPadding int32
}

/****************************************************************************
 * Encoder parameters
 ****************************************************************************/
/* CPU flags */

const (
	/* x86 */
	CpuMmx    uint32 = 1 << 0
	CpuMmx2   uint32 = 1 << 1 /* MMX2 aka MMXEXT aka ISSE */
	CpuMmxext        = CpuMmx2
	CpuSse    uint32 = 1 << 2
	CpuSse2   uint32 = 1 << 3
	CpuLzcnt  uint32 = 1 << 4
	CpuSse3   uint32 = 1 << 5
	CpuSsse3  uint32 = 1 << 6
	CpuSse4   uint32 = 1 << 7  /* SSE4.1 */
	CpuSse42  uint32 = 1 << 8  /* SSE4.2 */
	CpuAvx    uint32 = 1 << 9  /* Requires OS support even if YMM registers aren't used */
	CpuXop    uint32 = 1 << 10 /* AMD XOP */
	CpuFma4   uint32 = 1 << 11 /* AMD FMA4 */
	CpuFma3   uint32 = 1 << 12
	CpuBmi1   uint32 = 1 << 13
	CpuBmi2   uint32 = 1 << 14
	CpuAvx2   uint32 = 1 << 15
	CpuAvx512 uint32 = 1 << 16 /* AVX-512 {F, CD, BW, DQ, VL}, requires OS support */
	/* x86 modifiers */
	CpuCacheline32 uint32 = 1 << 17 /* avoid memory loads that span the border between two cachelines */
	CpuCacheline64 uint32 = 1 << 18 /* 32/64 is the size of a cacheline in bytes */
	CpuSse2IsSlow  uint32 = 1 << 19 /* avoid most SSE2 functions on Athlon64 */
	CpuSse2IsFast  uint32 = 1 << 20 /* a few functions are only faster on Core2 and Phenom */
	CpuSlowShuffle uint32 = 1 << 21 /* The Conroe has a slow shuffle unit (relative to overall SSE performance) */
	CpuStackMod4   uint32 = 1 << 22 /* if stack is only mod4 and not mod16 */
	CpuSlowAtom    uint32 = 1 << 23 /* The Atom is terrible: slow SSE unaligned loads, slow
	 * SIMD multiplies, slow SIMD variable shifts, slow pshufb,
	 * cacheline split penalties -- gather everything here that
	 * isn't shared by other CPUs to avoid making half a dozen
	 * new SLOW flags. */
	CpuSlowPshufb  uint32 = 1 << 24 /* such as on the Intel Atom */
	CpuSlowPalignr uint32 = 1 << 25 /* such as on the AMD Bobcat */

	/* PowerPC */
	CpuAltivec uint32 = 0x0000001

	/* ARM and AArch64 */
	CpuArmv6       uint32 = 0x0000001
	CpuNeon        uint32 = 0x0000002 /* ARM NEON */
	CpuFastNeonMrc uint32 = 0x0000004 /* Transfer from NEON to ARM register is fast (Cortex-A9) */
	CpuArmv8       uint32 = 0x0000008

	/* MIPS */
	CpuMsa uint32 = 0x0000001 /* MIPS MSA */

	/* Analyse flags */
	AnalyseI4x4      uint32 = 0x0001 /* Analyse i4x4 */
	AnalyseI8x8      uint32 = 0x0002 /* Analyse i8x8 (requires 8x8 transform) */
	AnalysePsub16x16 uint32 = 0x0010 /* Analyse p16x8, p8x16 and p8x8 */
	AnalysePsub8x8   uint32 = 0x0020 /* Analyse p8x4, p4x8, p4x4 */
	AnalyseBsub16x16 uint32 = 0x0100 /* Analyse b16x8, b8x16 and b8x8 */

	DirectPredNone       = 0
	DirectPredSpatial    = 1
	DirectPredTemporal   = 2
	DirectPredAuto       = 3
	MeDia                = 0
	MeHex                = 1
	MeUmh                = 2
	MeEsa                = 3
	MeTesa               = 4
	CqmFlat              = 0
	CqmJvt               = 1
	CqmCustom            = 2
	RcCqp                = 0
	RcCrf                = 1
	RcAbr                = 2
	QpAuto               = 0
	AqNone               = 0
	AqVariance           = 1
	AqAutovariance       = 2
	AqAutovarianceBiased = 3
	BAdaptNone           = 0
	BAdaptFast           = 1
	BAdaptTrellis        = 2
	WeightpNone          = 0
	WeightpSimple        = 1
	WeightpSmart         = 2
	BPyramidNone         = 0
	BPyramidStrict       = 1
	BPyramidNormal       = 2
	KeyintMinAuto        = 0
	KeyintMaxInfinite    = 1 << 30

	/* AVC-Intra flavors */
	AvcintraFlavorPanasonic = 0
	AvcintraFlavorSony      = 1

	/* !to add missing names */
	/* static const char * const x264_direct_pred_names[] = { "none", "spatial", "temporal", "auto", 0 }; */
	/* static const char * const x264_motion_est_names[] = { "dia", "hex", "umh", "esa", "tesa", 0 }; */
	/* static const char * const x264_b_pyramid_names[] = { "none", "strict", "normal", 0 }; */
	/* static const char * const x264_overscan_names[] = { "undef", "show", "crop", 0 }; */
	/* static const char * const x264_vidformat_names[] = { "component", "pal", "ntsc", "secam", "mac", "undef", 0 }; */
	/* static const char * const x264_fullrange_names[] = { "off", "on", 0 }; */
	/* static const char * const x264_colorprim_names[] = { "", "bt709", "undef", "", "bt470m", "bt470bg", "smpte170m", "smpte240m", "film", "bt2020", "smpte428", "smpte431", "smpte432", 0 }; */
	/* static const char * const x264_transfer_names[] = { "", "bt709", "undef", "", "bt470m", "bt470bg", "smpte170m", "smpte240m", "linear", "log100", "log316", "iec61966-2-4", "bt1361e", "iec61966-2-1", "bt2020-10", "bt2020-12", "smpte2084", "smpte428", "arib-std-b67", 0 }; */
	/* static const char * const x264_colmatrix_names[] = { "GBR", "bt709", "undef", "", "fcc", "bt470bg", "smpte170m", "smpte240m", "YCgCo", "bt2020nc", "bt2020c", "smpte2085", "chroma-derived-nc", "chroma-derived-c", "ICtCp", 0 }; */
	/* static const char * const x264_nal_hrd_names[] = { "none", "vbr", "cbr", 0 }; */
	/* static const char * const x264_avcintra_flavor_names[] = { "panasonic", "sony", 0 }; */

	/* Colorspace type */
	CspMask      = 0x00ff /* */
	CspNone      = 0x0000 /* Invalid mode     */
	CspI400      = 0x0001 /* monochrome 4:0:0 */
	CspI420      = 0x0002 /* yuv 4:2:0 planar */
	CspYv12      = 0x0003 /* yvu 4:2:0 planar */
	CspNv12      = 0x0004 /* yuv 4:2:0, with one y plane and one packed u+v */
	CspNv21      = 0x0005 /* yuv 4:2:0, with one y plane and one packed v+u */
	CspI422      = 0x0006 /* yuv 4:2:2 planar */
	CspYv16      = 0x0007 /* yvu 4:2:2 planar */
	CspNv16      = 0x0008 /* yuv 4:2:2, with one y plane and one packed u+v */
	CspYuyv      = 0x0009 /* yuyv 4:2:2 packed */
	CspUyvy      = 0x000a /* uyvy 4:2:2 packed */
	CspV210      = 0x000b /* 10-bit yuv 4:2:2 packed in 32 */
	CspI444      = 0x000c /* yuv 4:4:4 planar */
	CspYv24      = 0x000d /* yvu 4:4:4 planar */
	CspBgr       = 0x000e /* packed bgr 24bits */
	CspBgra      = 0x000f /* packed bgr 32bits */
	CspRgb       = 0x0010 /* packed rgb 24bits */
	CspMax       = 0x0011 /* end of list */
	CspVflip     = 0x1000 /* the csp is vertically flipped */
	CspHighDepth = 0x2000 /* the csp has a depth of 16 bits per pixel component */

	/* Slice type */
	TypeAuto     = 0x0000 /* Let x264 choose the right type */
	TypeIdr      = 0x0001
	TypeI        = 0x0002
	TypeP        = 0x0003
	TypeBref     = 0x0004 /* Non-disposable B-frame */
	TypeB        = 0x0005
	TypeKeyframe = 0x0006 /* IDR or I depending on b_open_gop option */
	/* !to reimplement macro */
	/* #define IS_X264_TYPE_I(x) ((x)==X264_TYPE_I || (x)==X264_TYPE_IDR || (x)==X264_TYPE_KEYFRAME) */
	/* #define IS_X264_TYPE_B(x) ((x)==X264_TYPE_B || (x)==X264_TYPE_BREF) */

	/* Log level */
	LogNone    = -1
	LogError   = 0
	LogWarning = 1
	LogInfo    = 2
	LogDebug   = 3

	/* Threading */
	ThreadsAuto       = 0  /* Automatically select optimal number of threads */
	SyncLookaheadAuto = -1 /* Automatically select optimal lookahead thread buffer size */

	/* HRD */
	NalHrdNone = 0
	NalHrdVbr  = 1
	NalHrdCbr  = 2
)

const (
	/* The macroblock is constant and remains unchanged from the previous frame. */
	MbinfoConstant = 1 << 0
	/* More flags may be added in the future. */
)

/* Zones: override ratecontrol or other options for specific sections of the video.
 * See x264_encoder_reconfig() for which options can be changed.
 * If zones overlap, whichever comes later in the list takes precedence. */
type Zone struct {
	IStart, IEnd   int32 /* range of frame numbers */
	BForceQp       int32 /* whether to use qp vs bitrate factor */
	IQp            int32
	FBitrateFactor float32
	Param          *Param
}

// Param????????????: https://www.jianshu.com/p/b46a33dd958d.
type Param struct {
	/* CPU flags */
	Cpu               uint32
	IThreads          int32 /* encode multiple frames in parallel */
	ILookaheadThreads int32 /* multiple threads for lookahead analysis */
	BSlicedThreads    int32 /* Whether to use slice-based threading. */
	BDeterministic    int32 /* whether to allow non-deterministic optimizations when threaded */
	BCpuIndependent   int32 /* force canonical behavior rather than cpu-dependent optimal algorithms */
	ISyncLookahead    int32 /* threaded lookahead buffer */

	/* Video Properties */
	IWidth      int32
	IHeight     int32
	ICsp        int32 /* CSP of encoded bitstream */
	IBitdepth   int32
	ILevelIdc   int32
	IFrameTotal int32 /* number of frames to encode if known, else 0 */

	/* NAL HRD
	 * Uses Buffering and Picture Timing SEIs to signal HRD
	 * The HRD in H.264 was not designed with VFR in mind.
	 * It is therefore not recommended to use NAL HRD with VFR.
	 * Furthermore, reconfiguring the VBV (via x264_encoder_reconfig)
	 * will currently generate invalid HRD. */
	INalHrd int32

	Vui struct {
		/* they will be reduced to be 0 < x <= 65535 and prime */
		ISarHeight int32
		ISarWidth  int32

		IOverscan int32 /* 0=undef, 1=no overscan, 2=overscan */

		/* see h264 annex E for the values of the following */
		IVidformat int32
		BFullrange int32
		IColorprim int32
		ITransfer  int32
		IColmatrix int32
		IChromaLoc int32 /* both top & bottom */
	}

	/* Bitstream parameters */
	IFrameReference int32 /* Maximum number of reference frames */
	IDpbSize        int32 /* Force a DPB size larger than that implied by B-frames and reference frames.
	 * Useful in combination with interactive error resilience. */

	// i?????????p?????????.
	IKeyintMax         int32 /* Force an IDR keyframe at this interval */
	IKeyintMin         int32 /* Scenecuts closer together than this are coded as I, not IDR. */
	IScenecutThreshold int32 /* how aggressively to insert extra I frames */
	BIntraRefresh      int32 /* Whether or not to use periodic intra refresh instead of IDR frames. */

	IBframe         int32 /* how many b-frame between 2 references pictures */
	IBframeAdaptive int32
	IBframeBias     int32
	IBframePyramid  int32 /* Keep some B-frames as references: 0=off, 1=strict hierarchical, 2=normal */
	BOpenGop        int32
	BBlurayCompat   int32
	IAvcintraClass  int32
	IAvcintraFlavor int32

	BDeblockingFilter        int32
	IDeblockingFilterAlphac0 int32 /* [-6, 6] -6 light filter, 6 strong */
	IDeblockingFilterBeta    int32 /* [-6, 6]  idem */

	BCabac        int32
	ICabacInitIdc int32

	BInterlaced       int32
	BConstrainedIntra int32

	ICqmPreset int32
	PszCqmFile *int8    /* filename (in UTF-8) of CQM file, JM format */
	Cqm4iy     [16]byte /* used only if i_cqm_preset == X264_CQM_CUSTOM */
	Cqm4py     [16]byte
	Cqm4ic     [16]byte
	Cqm4pc     [16]byte
	Cqm8iy     [64]byte
	Cqm8py     [64]byte
	Cqm8ic     [64]byte
	Cqm8pc     [64]byte

	/* Log */
	PfLog       *[0]byte
	PLogPrivate unsafe.Pointer
	ILogLevel   int32
	BFullRecon  int32 /* fully reconstruct frames, even when not necessary for encoding.  Implied by psz_dump_yuv */
	PszDumpYuv  *int8 /* filename (in UTF-8) for reconstructed frames */

	/* Encoder analyser parameters */
	Analyse struct {
		Intra uint32 /* intra partitions */
		Inter uint32 /* inter partitions */

		BTransform8x8   int32
		IWeightedPred   int32 /* weighting for P-frames */
		BWeightedBipred int32 /* implicit weighting for B-frames */
		IDirectMvPred   int32 /* spatial vs temporal mv prediction */
		IChromaQpOffset int32

		IMeMethod        int32   /* motion estimation algorithm to use (X264_ME_*) */
		IMeRange         int32   /* integer pixel motion estimation search range (from predicted mv) */
		IMvRange         int32   /* maximum length of a mv (in pixels). -1 = auto, based on level */
		IMvRangeThread   int32   /* minimum space between threads. -1 = auto, based on number of threads. */
		ISubpelRefine    int32   /* subpixel motion estimation quality */
		BChromaMe        int32   /* chroma ME for subpel and mode decision in P-frames */
		BMixedReferences int32   /* allow each mb partition to have its own reference number */
		ITrellis         int32   /* trellis RD quantization */
		BFastPskip       int32   /* early SKIP detection on P-frames */
		BDctDecimate     int32   /* transform coefficient thresholding on P-frames */
		INoiseReduction  int32   /* adaptive pseudo-deadzone */
		FPsyRd           float32 /* Psy RD strength */
		FPsyTrellis      float32 /* Psy trellis strength */
		BPsy             int32   /* Toggle all psy optimizations */

		BMbInfo       int32 /* Use input mb_info data in x264_picture_t */
		BMbInfoUpdate int32 /* Update the values in mb_info according to the results of encoding. */

		/* the deadzone size that will be used in luma quantization */
		ILumaDeadzone [2]int32

		BPsnr int32 /* compute and print PSNR stats */
		BSsim int32 /* compute and print SSIM stats */
	}

	/* Rate control parameters */
	Rc struct {
		IRcMethod int32 /* X264_RC_* */

		IQpConstant int32 /* 0=lossless */
		IQpMin      int32 /* min allowed QP value */
		IQpMax      int32 /* max allowed QP value */
		IQpStep     int32 /* max QP step between frames */

		IBitrate       int32
		FRfConstant    float32 /* 1pass VBR, nominal QP */
		FRfConstantMax float32 /* In CRF mode, maximum CRF as caused by VBV */
		FRateTolerance float32
		IVbvMaxBitrate int32
		IVbvBufferSize int32
		FVbvBufferInit float32 /* <=1: fraction of buffer_size. >1: kbit */
		FIpFactor      float32
		FPbFactor      float32

		/* VBV filler: force CBR VBV and use filler bytes to ensure hard-CBR.
		 * Implied by NAL-HRD CBR. */
		BFiller int32

		IAqMode     int32 /* psy adaptive QP. (X264_AQ_*) */
		FAqStrength float32
		BMbTree     int32 /* Macroblock-tree ratecontrol. */
		ILookahead  int32

		/* 2pass */
		BStatWrite int32 /* Enable stat writing in psz_stat_out */
		PszStatOut *int8 /* output filename (in UTF-8) of the 2pass stats file */
		BStatRead  int32 /* Read stat from psz_stat_in and use it */
		PszStatIn  *int8 /* input filename (in UTF-8) of the 2pass stats file */

		/* 2pass params (same as ffmpeg ones) */
		FQcompress      float32 /* 0.0 => cbr, 1.0 => constant qp */
		FQblur          float32 /* temporally blur quants */
		FComplexityBlur float32 /* temporally blur complexity */
		Zones           *Zone   /* ratecontrol overrides */
		IZones          int32   /* number of zone_t's */
		PszZones        *int8   /* alternate method of specifying zones */
	}

	/* Cropping Rectangle parameters: added to those implicitly defined by
	   non-mod16 video resolutions. */
	CropRect struct {
		ILeft   int32
		ITop    int32
		IRight  int32
		IBottom int32
	}

	/* frame packing arrangement flag */
	IFramePacking int32

	/* alternative transfer SEI */
	IAlternativeTransfer int32

	/* Muxing parameters */
	BAud           int32 /* generate access unit delimiters */
	BRepeatHeaders int32 /* put SPS/PPS before each keyframe */
	BAnnexb        int32 /* if set, place start codes (4 bytes) before NAL units,
	 * otherwise place size (4 bytes) before NAL units. */
	ISpsId    int32 /* SPS and PPS id number */
	BVfrInput int32 /* VFR input.  If 1, use timebase and timestamps for ratecontrol purposes.
	 * If 0, use fps only. */
	BPulldown    int32  /* use explicity set timebase for CFR */
	IFpsNum      uint32 // 25 default
	IFpsDen      uint32 // 1 default
	ITimebaseNum uint32 /* Timebase numerator */
	ITimebaseDen uint32 /* Timebase denominator */

	BTff int32

	/* Pulldown:
	 * The correct pic_struct must be passed with each input frame.
	 * The input timebase should be the timebase corresponding to the output framerate. This should be constant.
	 * e.g. for 3:2 pulldown timebase should be 1001/30000
	 * The PTS passed with each frame must be the PTS of the frame after pulldown is applied.
	 * Frame doubling and tripling require b_vfr_input set to zero (see H.264 Table D-1)
	 *
	 * Pulldown changes are not clearly defined in H.264. Therefore, it is the calling app's responsibility to manage this.
	 */

	BPicStruct int32

	/* Fake Interlaced.
	 *
	 * Used only when b_interlaced=0. Setting this flag makes it possible to flag the stream as PAFF interlaced yet
	 * encode all frames progressively. It is useful for encoding 25p and 30p Blu-Ray streams.
	 */
	BFakeInterlaced int32

	/* Don't optimize header parameters based on video content, e.g. ensure that splitting an input video, compressing
	 * each part, and stitching them back together will result in identical SPS/PPS. This is necessary for stitching
	 * with container formats that don't allow multiple SPS/PPS. */
	BStitchable int32

	BOpencl        int32          /* use OpenCL when available */
	IOpenclDevice  int32          /* specify count of GPU devices to skip, for CLI users */
	OpenclDeviceId unsafe.Pointer /* pass explicit cl_device_id as void*, for API users */
	PszClbinFile   *int8          /* filename (in UTF-8) of the compiled OpenCL kernel cache file */

	/* Slicing parameters */
	iSliceMaxSize  int32 /* Max size per slice in bytes; includes estimated NAL overhead. */
	iSliceMaxMbs   int32 /* Max number of MBs per slice; overrides iSliceCount. */
	iSliceMinMbs   int32 /* Min number of MBs per slice */
	iSliceCount    int32 /* Number of slices per frame: forces rectangular slices. */
	iSliceCountMax int32 /* Absolute cap on slices per frame; stops applying slice-max-size
	 * and slice-max-mbs if this is reached. */

	ParamFree   *func(arg unsafe.Pointer)
	NaluProcess *func(H []T, Nal []Nal, Opaque unsafe.Pointer)

	Opaque unsafe.Pointer
}

/****************************************************************************
 * H.264 level restriction information
 ****************************************************************************/

type Level struct {
	LevelIdc  byte
	Mbps      int32  /* max macroblock processing rate (macroblocks/sec) */
	FrameSize int32  /* max frame size (macroblocks) */
	Dpb       int32  /* max decoded picture buffer (mbs) */
	Bitrate   int32  /* max bitrate (kbit/sec) */
	Cpb       int32  /* max vbv buffer (kbit) */
	MvRange   uint16 /* max vertical mv component range (pixels) */
	MvsPer2mb byte   /* max mvs per 2 consecutive mbs. */
	SliceRate byte   /* ?? */
	Mincr     byte   /* min compression ratio */
	Bipred8x8 byte   /* limit bipred to >=8x8 */
	Direct8x8 byte   /* limit b_direct to >=8x8 */
	FrameOnly byte   /* forbid interlacing */
}

type PicStruct int32

const (
	PicStructAuto        = iota // automatically decide (default)
	PicStructProgressive = 1    // progressive frame
	// "TOP" and "BOTTOM" are not supported in x264 (PAFF only)
	PicStructTopBottom       = 4 // top field followed by bottom
	PicStructBottomTop       = 5 // bottom field followed by top
	PicStructTopBottomTop    = 6 // top field, bottom field, top field repeated
	PicStructBottomTopBottom = 7 // bottom field, top field, bottom field repeated
	PicStructDouble          = 8 // double frame
	PicStructTriple          = 9 // triple frame
)

type Hrd struct {
	CpbInitialArrivalTime float64
	CpbFinalArrivalTime   float64
	CpbRemovalTime        float64

	DpbOutputTime float64
}

/* Arbitrary user SEI:
 * Payload size is in bytes and the payload pointer must be valid.
 * Payload types and syntax can be found in Annex D of the H.264 Specification.
 * SEI payload alignment bits as described in Annex D must be included at the
 * end of the payload if needed.
 * The payload should not be NAL-encapsulated.
 * Payloads are written first in order of input, apart from in the case when HRD
 * is enabled where payloads are written after the Buffering Period SEI. */
type SeiPayload struct {
	PayloadSize int32
	PayloadType int32
	Payload     *byte
}

type Sei struct {
	NumPayloads int32
	Payloads    *SeiPayload
	/* In: optional callback to free each payload AND x264_sei_payload_t when used. */
	SeiFree *func(arg0 unsafe.Pointer)
}

type Image struct {
	ICsp    int32             /* Colorspace */
	IPlane  int32             /* Number of image planes */
	IStride [4]int32          /* Strides for each plane */
	Plane   [4]unsafe.Pointer /* Pointers to each plane */
}

/* All arrays of data here are ordered as follows:
 * each array contains one offset per macroblock, in raster scan order.  In interlaced
 * mode, top-field MBs and bottom-field MBs are interleaved at the row level.
 * Macroblocks are 16x16 blocks of pixels (with respect to the luma plane).  For the
 * purposes of calculating the number of macroblocks, width and height are rounded up to
 * the nearest 16.  If in interlaced mode, height is rounded up to the nearest 32 instead. */
type ImageProperties struct {
	/* In: an array of quantizer offsets to be applied to this image during encoding.
	 *     These are added on top of the decisions made by x264.
	 *     Offsets can be fractional; they are added before QPs are rounded to integer.
	 *     Adaptive quantization must be enabled to use this feature.  Behavior if quant
	 *     offsets differ between encoding passes is undefined. */
	QuantOffsets *float32
	/* In: optional callback to free quant_offsets when used.
	*     Useful if one wants to use a different quant_offset array for each frame. */
	QuantOffsetsFree *func(arg0 unsafe.Pointer)

	/* In: optional array of flags for each macroblock.
	 *     Allows specifying additional information for the encoder such as which macroblocks
	 *     remain unchanged.  Usable flags are listed below.
	 *     x264_param_t.analyse.b_mb_info must be set to use this, since x264 needs to track
	 *     extra data internally to make full use of this information.
	 *
	 * Out: if b_mb_info_update is set, x264 will update this array as a result of encoding.
	 *
	 *      For "MBINFO_CONSTANT", it will remove this flag on any macroblock whose decoded
	 *      pixels have changed.  This can be useful for e.g. noting which areas of the
	 *      frame need to actually be blitted. Note: this intentionally ignores the effects
	 *      of deblocking for the current frame, which should be fine unless one needs exact
	 *      pixel-perfect accuracy.
	 *
	 *      Results for MBINFO_CONSTANT are currently only set for P-frames, and are not
	 *      guaranteed to enumerate all blocks which haven't changed.  (There may be false
	 *      negatives, but no false positives.)
	 */
	MbInfo *byte
	/* In: optional callback to free mb_info when used. */
	MbInfoFree *func(arg0 unsafe.Pointer)

	/* Out: SSIM of the the frame luma (if x264_param_t.b_ssim is set) */
	FSsim float64
	/* Out: Average PSNR of the frame (if x264_param_t.b_psnr is set) */
	FPsnrAvg float64
	/* Out: PSNR of Y, U, and V (if x264_param_t.b_psnr is set) */
	FPsnr [3]float64

	/* Out: Average effective CRF of the encoded frame */
	FCrfAvg float64
}

//
type Picture struct {
	/* In: force picture type (if not auto)
	 *     If x264 encoding parameters are violated in the forcing of picture types,
	 *     x264 will correct the input picture type and log a warning.
	 * Out: type of the picture encoded */
	IType int32
	/* In: force quantizer for != X264_QP_AUTO */
	IQpplus1 int32
	/* In: pic_struct, for pulldown/doubling/etc...used only if b_pic_struct=1.
	 *     use pic_struct_e for pic_struct inputs
	 * Out: pic_struct element associated with frame */
	IPicStruct int32
	/* Out: whether this frame is a keyframe.  Important when using modes that result in
	 * SEI recovery points being used instead of IDR frames. */
	BKeyframe int32
	/* In: user pts, Out: pts of encoded picture (user)*/
	IPts int64
	/* Out: frame dts. When the pts of the first frame is close to zero,
	 *      initial frames may have a negative dts which must be dealt with by any muxer */
	IDts int64
	/* In: custom encoding parameters to be set from this frame forwards
	   (in coded order, not display order). If NULL, continue using
	   parameters from the previous frame.  Some parameters, such as
	   aspect ratio, can only be changed per-GOP due to the limitations
	   of H.264 itself; in this case, the caller must force an IDR frame
	   if it needs the changed parameter to apply immediately. */
	Param *Param
	/* In: raw image data */
	/* Out: reconstructed image data.  x264 may skip part of the reconstruction process,
	   e.g. deblocking, in frames where it isn't necessary.  To force complete
	   reconstruction, at a small speed cost, set b_full_recon. */
	Img Image
	/* In: optional information to modify encoder decisions for this frame
	 * Out: information about the encoded frame */
	Prop ImageProperties
	/* Out: HRD timing information. Output only when i_nal_hrd is set. */
	Hrdiming Hrd
	/* In: arbitrary user SEI (e.g subtitles, AFDs) */
	ExtraSei Sei
	/* private user data. copied from input to output frames. */
	Opaque unsafe.Pointer
}

func (p *Picture) freePlane(n int) {
	C.free(p.Img.Plane[n])
}

func (t *T) cptr() *C.x264_t { return (*C.x264_t)(unsafe.Pointer(t)) }

func (n *Nal) cptr() *C.x264_nal_t { return (*C.x264_nal_t)(unsafe.Pointer(n)) }

func (p *Param) cptr() *C.x264_param_t { return (*C.x264_param_t)(unsafe.Pointer(p)) }

func (p *Picture) cptr() *C.x264_picture_t { return (*C.x264_picture_t)(unsafe.Pointer(p)) }

// NalEncode - encode Nal.
func NalEncode(h *T, dst []byte, nal *Nal) {
	ch := h.cptr()
	cdst := (*C.uint8_t)(unsafe.Pointer(&dst[0]))
	cnal := nal.cptr()
	C.x264_nal_encode(ch, cdst, cnal)
}

// ParamDefault - fill Param with default values and do CPU detection.
func ParamDefault(param *Param) {
	C.x264_param_default(param.cptr())
}

// ParamParse - set one parameter by name. Returns 0 on success.
func ParamParse(param *Param, name string, value string) int32 {
	cparam := param.cptr()

	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))

	cvalue := C.CString(value)
	defer C.free(unsafe.Pointer(cvalue))

	ret := C.x264_param_parse(cparam, cname, cvalue)
	v := (int32)(ret)
	return v
}

// ParamDefaultPreset - the same as ParamDefault, but also use the passed preset and tune to modify the default settings
// (either can be nil, which implies no preset or no tune, respectively).
//
// Currently available presets are, ordered from fastest to slowest:
// "ultrafast", "superfast", "veryfast", "faster", "fast", "medium", "slow", "slower", "veryslow", "placebo".
//
// Currently available tunings are:
// "film", "animation", "grain", "stillimage", "psnr", "ssim", "fastdecode", "zerolatency".
//
// Returns 0 on success, negative on failure (e.g. invalid preset/tune name).
func ParamDefaultPreset(param *Param, preset string, tune string) int32 {
	cparam := param.cptr()

	cpreset := C.CString(preset)
	defer C.free(unsafe.Pointer(cpreset))

	ctune := C.CString(tune)
	defer C.free(unsafe.Pointer(ctune))

	ret := C.x264_param_default_preset(cparam, cpreset, ctune)
	v := (int32)(ret)
	return v
}

// ParamApplyFastfirstpass - if first-pass mode is set (rc.b_stat_read == 0, rc.b_stat_write == 1),
// modify the encoder settings to disable options generally not useful on the first pass.
func ParamApplyFastfirstpass(param *Param) {
	cparam := param.cptr()
	C.x264_param_apply_fastfirstpass(cparam)
}

// ParamApplyProfile - applies the restrictions of the given profile.
//
// Currently available profiles are, from most to least restrictive:
// "baseline", "main", "high", "high10", "high422", "high444".
// (can be nil, in which case the function will do nothing).
//
// Returns 0 on success, negative on failure (e.g. invalid profile name).
func ParamApplyProfile(param *Param, profile string) int32 {
	cparam := param.cptr()

	cprofile := C.CString(profile)
	defer C.free(unsafe.Pointer(cprofile))

	ret := C.x264_param_apply_profile(cparam, cprofile)
	v := (int32)(ret)
	return v
}

// PictureInit - initialize an Picture. Needs to be done if the calling application
// allocates its own Picture as opposed to using PictureAlloc.
func PictureInit(pic *Picture) {
	cpic := pic.cptr()
	C.x264_picture_init(cpic)
}

// PictureAlloc - alloc data for a Picture. You must call PictureClean on it.
// Returns 0 on success, or -1 on malloc failure or invalid colorspace.
func PictureAlloc(pic *Picture, iCsp int32, iWidth int32, iHeight int32) int32 {
	cpic := pic.cptr()

	ciCsp := (C.int)(iCsp)
	ciWidth := (C.int)(iWidth)
	ciHeight := (C.int)(iHeight)

	ret := C.x264_picture_alloc(cpic, ciCsp, ciWidth, ciHeight)
	v := (int32)(ret)
	return v
}

// PictureClean - free associated resource for a Picture allocated with PictureAlloc ONLY.
func PictureClean(pic *Picture) {
	cpic := pic.cptr()
	C.x264_picture_clean(cpic)
}

// EncoderOpen - create a new encoder handler, all parameters from Param are copied.
func EncoderOpen(param *Param) *T {
	cparam := param.cptr()

	ret := C.x264_encoder_open(cparam)
	v := *(**T)(unsafe.Pointer(&ret))
	return v
}

// EncoderReconfig - various parameters from Param are copied.
// Returns 0 on success, negative on parameter validation error.
func EncoderReconfig(enc *T, param *Param) int32 {
	cenc := enc.cptr()
	cparam := param.cptr()

	ret := C.x264_encoder_reconfig(cenc, cparam)
	v := (int32)(ret)
	return v
}

// EncoderParameters - copies the current internal set of parameters to the pointer provided.
func EncoderParameters(enc *T, param *Param) {
	cenc := enc.cptr()
	cparam := param.cptr()

	C.x264_encoder_parameters(cenc, cparam)
}

// EncoderHeaders - return the SPS and PPS that will be used for the whole stream.
// Returns the number of bytes in the returned NALs or negative on error.
func EncoderHeaders(enc *T, ppNal []*Nal, piNal *int32) int32 {
	cenc := enc.cptr()

	cppNal := (**C.x264_nal_t)(unsafe.Pointer(&ppNal[0]))
	cpiNal := (*C.int)(unsafe.Pointer(piNal))

	ret := C.x264_encoder_headers(cenc, cppNal, cpiNal)
	v := (int32)(ret)
	return v
}

// EncoderEncode - encode one picture.
// Returns the number of bytes in the returned NALs, negative on error and zero if no NAL units returned.
// ????????????: https://blog.csdn.net/liuzhihan209/article/details/8959494.
func EncoderEncode(enc *T, ppNal []*Nal, piNal *int32, picIn *Picture, picOut *Picture) int32 {
	cenc := enc.cptr()

	cppNal := (**C.x264_nal_t)(unsafe.Pointer(&ppNal[0]))
	cpiNal := (*C.int)(unsafe.Pointer(piNal))

	cpicIn := picIn.cptr()
	cpicOut := picOut.cptr()

	ret := C.x264_encoder_encode(cenc, cppNal, cpiNal, cpicIn, cpicOut)
	v := (int32)(ret)
	return v
}

// EncoderClose - close an encoder handler.
func EncoderClose(enc *T) {
	cenc := enc.cptr()
	C.x264_encoder_close(cenc)
}

// EncoderDelayedFrames - return the number of currently delayed (buffered) frames.
// This should be used at the end of the stream, to know when you have all the encoded frames.
func EncoderDelayedFrames(enc *T) int32 {
	cenc := enc.cptr()

	ret := C.x264_encoder_delayed_frames(cenc)
	v := (int32)(ret)
	return v
}

// EncoderMaximumDelayedFrames - return the maximum number of delayed (buffered) frames that can occur with the current parameters.
func EncoderMaximumDelayedFrames(enc *T) int32 {
	cenc := enc.cptr()

	ret := C.x264_encoder_maximum_delayed_frames(cenc)
	v := (int32)(ret)
	return v
}

// EncoderIntraRefresh - If an intra refresh is not in progress, begin one with the next P-frame.
// If an intra refresh is in progress, begin one as soon as the current one finishes.
// Requires that BIntraRefresh be set.
//
// Should not be called during an x264_encoder_encode.
func EncoderIntraRefresh(enc *T) {
	cenc := enc.cptr()
	C.x264_encoder_intra_refresh(cenc)
}

// EncoderInvalidateReference - An interactive error resilience tool, designed for use in a low-latency one-encoder-few-clients system.
// Should not be called during an EncoderEncode, but multiple calls can be made simultaneously.
//
// Returns 0 on success, negative on failure.
func EncoderInvalidateReference(enc *T, pts int) int32 {
	cenc := enc.cptr()
	cpts := (C.int64_t)(pts)

	ret := C.x264_encoder_invalidate_reference(cenc, cpts)
	v := (int32)(ret)
	return v
}
