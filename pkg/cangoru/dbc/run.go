package dbc

import (
	"log"
	"strings"
)

func Test() {
	scanner := newScanner(strings.NewReader(file1), file1)
	for {
		token := scanner.scan()
		log.Print(token.line, ":", token.col, " ", token.kindName, " ", token.value)
		if token.kind == tokenEOF || token.kind == tokenError {
			break
		}
	}
}

var file = `VERSION "1.0"

NS_ : 
	NS_DESC_
	CM_
	BA_DEF_
	BA_
	VAL_
	CAT_DEF_
	CAT_
	FILTER
	BA_DEF_DEF_
	EV_DATA_
	ENVVAR_DATA_
	SGTYPE_
	SGTYPE_VAL_
	BA_DEF_SGTYPE_
	BA_SGTYPE_
	SIG_TYPE_REF_
	VAL_TABLE_
	SIG_GROUP_
	SIG_VALTYPE_
	SIGTYPE_VALTYPE_
	BO_TX_BU_
	BA_DEF_REL_
	BA_REL_
	BA_DEF_DEF_REL_
	BU_SG_REL_
	BU_EV_REL_
	BU_BO_REL_
	SG_MUL_VAL_

BS_ :

BU_ : Node1 Node2

BO_ 2364540158 EEC1: 8 Vector__XXX
	SG_ Engine_Status : 0|2@1+ (0,0) [0|2] "status" Vector__XXX
	SG_ Engine_Speed : 24|16@0+ (0.125,0) [0|8031.875] "rpm" Node1,Node2


CM_ BU_ Node1 "Node1 desc";
CM_ BU_ Node2 "Node2 desc";
CM_ BO_ 2364540158 "desc 0 (period: 100 ms)";
CM_ SG_ 2364540158 Engine_Speed "desc 1";
CM_ SG_ 2364540158 Engine_Status "desc 2";

BA_DEF_ "FloatAtt" FLOAT 0 25.75;
BA_DEF_ BU_ "TestString" STRING ;
BA_DEF_ BO_ "VFrameFormat" ENUM "StandardCAN","ExtendedCAN","reserved","J1939PG";
BA_DEF_ BO_ "MsgPeriodMS" INT 0 65535;
BA_DEF_ SG_ "SPN" INT 0 524287;
BA_DEF_ SG_ "TestFloatAtt" FLOAT 0 25.75;
BA_DEF_DEF_ "FloatAtt" 1.5;
BA_DEF_DEF_ "TestString" "";
BA_DEF_DEF_ "VFrameFormat" "J1939PG";
BA_DEF_DEF_ "MsgPeriodMS" 0;
BA_DEF_DEF_ "SPN" 0;
BA_DEF_DEF_ "TestFloatAtt" 1.5;
BA_ "TestString" BU_ Node2 "test";
BA_ "MsgPeriodMS" BO_ 2364540158 100;
BA_ "VFrameFormat" BO_ 2364540158 0;
BA_ "SPN" SG_ 2364540158 Engine_Speed 190;
BA_ "TestFloatAtt" SG_ 2364540158 Engine_Speed 10.5;

VAL_ 2364540158 Engine_Status 0 "Off" 1 "Idle" 2 "Running";
`

var file1 = `VERSION "1.0"

NS_ : 
	NS_DESC_
	CM_
	BA_DEF_
	BA_
	VAL_
	CAT_DEF_
	CAT_
	FILTER
	BA_DEF_DEF_
	EV_DATA_
	ENVVAR_DATA_
	SGTYPE_
	SGTYPE_VAL_
	BA_DEF_SGTYPE_
	BA_SGTYPE_
	SIG_TYPE_REF_
	VAL_TABLE_
	SIG_GROUP_
	SIG_VALTYPE_
	SIGTYPE_VALTYPE_
	BO_TX_BU_
	BA_DEF_REL_
	BA_REL_
	BA_DEF_DEF_REL_
	BU_SG_REL_
	BU_EV_REL_
	BU_BO_REL_
	SG_MUL_VAL_

BS_ :

BU_ : Node1

BO_ 2024 OBD2: 8 Node1
	SG_ Service M : 11|4@1+ (1,0) [0|15] "" Vector__XXX
	SG_ ExtendedMuxSignalName m1M : 23|8@1+ (1,0) [0|255] "" Vector__XXX
	SG_ MultiplexedSignalName m2 : 23|8@1+ (1,0) [0|255] "" Vector__XXX
	SG_ VehicleSpeed m13 : 31|8@1+ (1,0) [0|255] "km/h" Vector__XXX
	SG_ ThrottlePosition m17 : 31|8@1+ (0.39216,0) [0|100] "%" Vector__XXX

CM_ BU_ Node1 "Node1 desc";
CM_ BO_ 2024 "desc OBD2";
CM_ SG_ 2024 Service "desc Service";
CM_ SG_ 2024 MultiplexedSignalName "desc S2";
CM_ SG_ 2024 ExtendedMuxSignalName "desc S1";
CM_ SG_ 2024 ThrottlePosition "desc ThrottlePosition";
CM_ SG_ 2024 VehicleSpeed "desc VehicleSpeed";

BA_DEF_ BO_ "MsgPeriodMS" INT 0 65535;
BA_DEF_ SG_ "TestFloatAtt" FLOAT 0 25.75;
BA_DEF_ SG_ "SPN" INT 0 524287;
BA_DEF_DEF_ "MsgPeriodMS" 0;
BA_DEF_DEF_ "SPN" 0;
BA_DEF_DEF_ "TestFloatAtt" 1.5;
BA_ "TestFloatAtt" SG_ 2024 VehicleSpeed 10.5;
BA_ "SPN" SG_ 2024 VehicleSpeed 190;

SG_MUL_VAL_ 2024 MultiplexedSignalName Service 2-2;
SG_MUL_VAL_ 2024 VehicleSpeed ExtendedMuxSignalName 13-13;
SG_MUL_VAL_ 2024 ThrottlePosition ExtendedMuxSignalName 17-17;
SG_MUL_VAL_ 2024 ExtendedMuxSignalName Service 1-1;
`
