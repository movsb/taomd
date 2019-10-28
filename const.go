package main

// Markdown punctuation.
//
// From https://spec.commonmark.org/0.29/#example-298
var punctuation = map[rune]int{
	'!':  1,
	'"':  1,
	'#':  1,
	'$':  1,
	'%':  1,
	'&':  1,
	'\'': 1,
	'(':  1,
	')':  1,
	'*':  1,
	'+':  1,
	',':  1,
	'-':  1,
	'.':  1,
	'/':  1,
	':':  1,
	';':  1,
	'<':  1,
	'=':  1,
	'>':  1,
	'?':  1,
	'@':  1,
	'[':  1,
	'\\': 1,
	']':  1,
	'^':  1,
	'_':  1,
	'`':  1,
	'{':  1,
	'|':  1,
	'}':  1,
	'~':  1,
}

// HTML5 Entity Name References
//
// Generated by tools/entities/entities.go,
// separated by blank lines to reduce file size.
//
// Non semilocon-ended names are not included.
// Finally, `&` and `;` are removed too.
//
// From https://html.spec.whatwg.org/entities.json
//
// TODO lazy initialize map?
var htmlEntities1 = map[string]rune{
	`ac`: 8766,
	`af`: 8289,
	`ap`: 8776,
	`DD`: 8517,
	`dd`: 8518,
	`ee`: 8519,
	`eg`: 10906,
	`el`: 10905,
	`gE`: 8807,
	`ge`: 8805,
	`gg`: 8811,
	`Gg`: 8921,
	`gl`: 8823,
	`gt`: 62,
	`GT`: 62,
	`Gt`: 8811,
	`ic`: 8291,
	`ii`: 8520,
	`Im`: 8465,
	`in`: 8712,
	`it`: 8290,
	`le`: 8804,
	`lE`: 8806,
	`lg`: 8822,
	`ll`: 8810,
	`Ll`: 8920,
	`Lt`: 8810,
	`LT`: 60,
	`lt`: 60,
	`mp`: 8723,
	`Mu`: 924,
	`mu`: 956,
	`ne`: 8800,
	`ni`: 8715,
	`nu`: 957,
	`Nu`: 925,
	`or`: 8744,
	`Or`: 10836,
	`oS`: 9416,
	`pi`: 960,
	`Pi`: 928,
	`pm`: 177,
	`Pr`: 10939,
	`pr`: 8826,
	`Re`: 8476,
	`rx`: 8478,
	`sc`: 8827,
	`Sc`: 10940,
	`wp`: 8472,
	`wr`: 8768,
	`Xi`: 926,
	`xi`: 958,

	`acd`: 8767,
	`acy`: 1072,
	`Acy`: 1040,
	`Afr`: 120068,
	`afr`: 120094,
	`amp`: 38,
	`AMP`: 38,
	`And`: 10835,
	`and`: 8743,
	`ang`: 8736,
	`apE`: 10864,
	`ape`: 8778,
	`ast`: 42,
	`Bcy`: 1041,
	`bcy`: 1073,
	`Bfr`: 120069,
	`bfr`: 120095,
	`bot`: 8869,
	`cap`: 8745,
	`Cap`: 8914,
	`cfr`: 120096,
	`Cfr`: 8493,
	`chi`: 967,
	`Chi`: 935,
	`cir`: 9675,
	`cup`: 8746,
	`Cup`: 8915,
	`Dcy`: 1044,
	`dcy`: 1076,
	`deg`: 176,
	`Del`: 8711,
	`dfr`: 120097,
	`Dfr`: 120071,
	`die`: 168,
	`div`: 247,
	`Dot`: 168,
	`dot`: 729,
	`ecy`: 1101,
	`Ecy`: 1069,
	`efr`: 120098,
	`Efr`: 120072,
	`egs`: 10902,
	`ell`: 8467,
	`els`: 10901,
	`eng`: 331,
	`ENG`: 330,
	`Eta`: 919,
	`eta`: 951,
	`eth`: 240,
	`ETH`: 208,
	`Fcy`: 1060,
	`fcy`: 1092,
	`Ffr`: 120073,
	`ffr`: 120099,
	`gap`: 10886,
	`gcy`: 1075,
	`Gcy`: 1043,
	`gel`: 8923,
	`gEl`: 10892,
	`geq`: 8805,
	`ges`: 10878,
	`Gfr`: 120074,
	`gfr`: 120100,
	`ggg`: 8921,
	`gla`: 10917,
	`glE`: 10898,
	`glj`: 10916,
	`gne`: 10888,
	`gnE`: 8809,
	`Hat`: 94,
	`hfr`: 120101,
	`Hfr`: 8460,
	`Icy`: 1048,
	`icy`: 1080,
	`iff`: 8660,
	`ifr`: 120102,
	`Ifr`: 8465,
	`int`: 8747,
	`Int`: 8748,
	`Jcy`: 1049,
	`jcy`: 1081,
	`Jfr`: 120077,
	`jfr`: 120103,
	`kcy`: 1082,
	`Kcy`: 1050,
	`kfr`: 120104,
	`Kfr`: 120078,
	`lap`: 10885,
	`lat`: 10923,
	`lcy`: 1083,
	`Lcy`: 1051,
	`lEg`: 10891,
	`leg`: 8922,
	`leq`: 8804,
	`les`: 10877,
	`Lfr`: 120079,
	`lfr`: 120105,
	`lgE`: 10897,
	`lnE`: 8808,
	`lne`: 10887,
	`loz`: 9674,
	`lrm`: 8206,
	`lsh`: 8624,
	`Lsh`: 8624,
	`Map`: 10501,
	`map`: 8614,
	`mcy`: 1084,
	`Mcy`: 1052,
	`Mfr`: 120080,
	`mfr`: 120106,
	`mho`: 8487,
	`mid`: 8739,
	`nap`: 8777,
	`ncy`: 1085,
	`Ncy`: 1053,
	`nfr`: 120107,
	`Nfr`: 120081,
	`nge`: 8817,
	`ngt`: 8815,
	`nis`: 8956,
	`niv`: 8715,
	`nle`: 8816,
	`nlt`: 8814,
	`Not`: 10988,
	`not`: 172,
	`npr`: 8832,
	`nsc`: 8833,
	`num`: 35,
	`ocy`: 1086,
	`Ocy`: 1054,
	`ofr`: 120108,
	`Ofr`: 120082,
	`ogt`: 10689,
	`ohm`: 937,
	`olt`: 10688,
	`ord`: 10845,
	`orv`: 10843,
	`par`: 8741,
	`Pcy`: 1055,
	`pcy`: 1087,
	`pfr`: 120109,
	`Pfr`: 120083,
	`phi`: 966,
	`Phi`: 934,
	`piv`: 982,
	`prE`: 10931,
	`pre`: 10927,
	`psi`: 968,
	`Psi`: 936,
	`Qfr`: 120084,
	`qfr`: 120110,
	`Rcy`: 1056,
	`rcy`: 1088,
	`REG`: 174,
	`reg`: 174,
	`rfr`: 120111,
	`Rfr`: 8476,
	`rho`: 961,
	`Rho`: 929,
	`rlm`: 8207,
	`Rsh`: 8625,
	`rsh`: 8625,
	`scE`: 10932,
	`sce`: 10928,
	`scy`: 1089,
	`Scy`: 1057,
	`Sfr`: 120086,
	`sfr`: 120112,
	`shy`: 173,
	`sim`: 8764,
	`smt`: 10922,
	`sol`: 47,
	`squ`: 9633,
	`Sub`: 8912,
	`sub`: 8834,
	`sum`: 8721,
	`Sum`: 8721,
	`sup`: 8835,
	`Sup`: 8913,
	`Tab`: 9,
	`Tau`: 932,
	`tau`: 964,
	`Tcy`: 1058,
	`tcy`: 1090,
	`Tfr`: 120087,
	`tfr`: 120113,
	`top`: 8868,
	`ucy`: 1091,
	`Ucy`: 1059,
	`Ufr`: 120088,
	`ufr`: 120114,
	`uml`: 168,
	`Vcy`: 1042,
	`vcy`: 1074,
	`Vee`: 8897,
	`vee`: 8744,
	`vfr`: 120115,
	`Vfr`: 120089,
	`wfr`: 120116,
	`Wfr`: 120090,
	`xfr`: 120117,
	`Xfr`: 120091,
	`Ycy`: 1067,
	`ycy`: 1099,
	`yen`: 165,
	`Yfr`: 120092,
	`yfr`: 120118,
	`zcy`: 1079,
	`Zcy`: 1047,
	`Zfr`: 8488,
	`zfr`: 120119,
	`zwj`: 8205,

	`andd`: 10844,
	`andv`: 10842,
	`ange`: 10660,
	`Aopf`: 120120,
	`aopf`: 120146,
	`apid`: 8779,
	`apos`: 39,
	`Ascr`: 119964,
	`ascr`: 119990,
	`Auml`: 196,
	`auml`: 228,
	`Barv`: 10983,
	`bbrk`: 9141,
	`Beta`: 914,
	`beta`: 946,
	`beth`: 8502,
	`bnot`: 8976,
	`bNot`: 10989,
	`bopf`: 120147,
	`Bopf`: 120121,
	`boxH`: 9552,
	`boxh`: 9472,
	`boxV`: 9553,
	`boxv`: 9474,
	`bscr`: 119991,
	`Bscr`: 8492,
	`bsim`: 8765,
	`bsol`: 92,
	`bull`: 8226,
	`bump`: 8782,
	`Cdot`: 266,
	`cdot`: 267,
	`cent`: 162,
	`CHcy`: 1063,
	`chcy`: 1095,
	`circ`: 710,
	`cirE`: 10691,
	`cire`: 8791,
	`comp`: 8705,
	`cong`: 8773,
	`copf`: 120148,
	`Copf`: 8450,
	`COPY`: 169,
	`copy`: 169,
	`cscr`: 119992,
	`Cscr`: 119966,
	`csub`: 10959,
	`csup`: 10960,
	`darr`: 8595,
	`Darr`: 8609,
	`dArr`: 8659,
	`dash`: 8208,
	`dHar`: 10597,
	`diam`: 8900,
	`DJcy`: 1026,
	`djcy`: 1106,
	`Dopf`: 120123,
	`dopf`: 120149,
	`Dscr`: 119967,
	`dscr`: 119993,
	`dscy`: 1109,
	`DScy`: 1029,
	`dsol`: 10742,
	`dtri`: 9663,
	`dzcy`: 1119,
	`DZcy`: 1039,
	`ecir`: 8790,
	`eDot`: 8785,
	`edot`: 279,
	`Edot`: 278,
	`emsp`: 8195,
	`ensp`: 8194,
	`Eopf`: 120124,
	`eopf`: 120150,
	`epar`: 8917,
	`epsi`: 949,
	`Escr`: 8496,
	`escr`: 8495,
	`Esim`: 10867,
	`esim`: 8770,
	`Euml`: 203,
	`euml`: 235,
	`euro`: 8364,
	`excl`: 33,
	`flat`: 9837,
	`fnof`: 402,
	`fopf`: 120151,
	`Fopf`: 120125,
	`fork`: 8916,
	`Fscr`: 8497,
	`fscr`: 119995,
	`gdot`: 289,
	`Gdot`: 288,
	`geqq`: 8807,
	`gjcy`: 1107,
	`GJcy`: 1027,
	`gnap`: 10890,
	`gneq`: 10888,
	`Gopf`: 120126,
	`gopf`: 120152,
	`Gscr`: 119970,
	`gscr`: 8458,
	`gsim`: 8819,
	`gtcc`: 10919,
	`half`: 189,
	`hArr`: 8660,
	`harr`: 8596,
	`hbar`: 8463,
	`Hopf`: 8461,
	`hopf`: 120153,
	`Hscr`: 8459,
	`hscr`: 119997,
	`Idot`: 304,
	`iecy`: 1077,
	`IEcy`: 1045,
	`imof`: 8887,
	`iocy`: 1105,
	`IOcy`: 1025,
	`iopf`: 120154,
	`Iopf`: 120128,
	`Iota`: 921,
	`iota`: 953,
	`Iscr`: 8464,
	`iscr`: 119998,
	`isin`: 8712,
	`iuml`: 239,
	`Iuml`: 207,
	`Jopf`: 120129,
	`jopf`: 120155,
	`jscr`: 119999,
	`Jscr`: 119973,
	`KHcy`: 1061,
	`khcy`: 1093,
	`kjcy`: 1116,
	`KJcy`: 1036,
	`Kopf`: 120130,
	`kopf`: 120156,
	`kscr`: 120000,
	`Kscr`: 119974,
	`Lang`: 10218,
	`lang`: 10216,
	`larr`: 8592,
	`lArr`: 8656,
	`Larr`: 8606,
	`late`: 10925,
	`lcub`: 123,
	`ldca`: 10550,
	`ldsh`: 8626,
	`leqq`: 8806,
	`lHar`: 10594,
	`LJcy`: 1033,
	`ljcy`: 1113,
	`lnap`: 10889,
	`lneq`: 10887,
	`Lopf`: 120131,
	`lopf`: 120157,
	`lozf`: 10731,
	`lpar`: 40,
	`lscr`: 120001,
	`Lscr`: 8466,
	`lsim`: 8818,
	`lsqb`: 91,
	`ltcc`: 10918,
	`ltri`: 9667,
	`macr`: 175,
	`male`: 9794,
	`malt`: 10016,
	`mlcp`: 10971,
	`mldr`: 8230,
	`Mopf`: 120132,
	`mopf`: 120158,
	`mscr`: 120002,
	`Mscr`: 8499,
	`nbsp`: 160,
	`ncap`: 10819,
	`ncup`: 10818,
	`ngeq`: 8817,
	`ngtr`: 8815,
	`nisd`: 8954,
	`NJcy`: 1034,
	`njcy`: 1114,
	`nldr`: 8229,
	`nleq`: 8816,
	`nmid`: 8740,
	`Nopf`: 8469,
	`nopf`: 120159,
	`npar`: 8742,
	`Nscr`: 119977,
	`nscr`: 120003,
	`nsim`: 8769,
	`nsub`: 8836,
	`nsup`: 8837,
	`ntgl`: 8825,
	`ntlg`: 8824,
	`oast`: 8859,
	`ocir`: 8858,
	`odiv`: 10808,
	`odot`: 8857,
	`ogon`: 731,
	`oint`: 8750,
	`omid`: 10678,
	`oopf`: 120160,
	`Oopf`: 120134,
	`opar`: 10679,
	`ordf`: 170,
	`ordm`: 186,
	`oror`: 10838,
	`oscr`: 8500,
	`Oscr`: 119978,
	`osol`: 8856,
	`ouml`: 246,
	`Ouml`: 214,
	`para`: 182,
	`part`: 8706,
	`perp`: 8869,
	`phiv`: 981,
	`plus`: 43,
	`Popf`: 8473,
	`popf`: 120161,
	`prap`: 10935,
	`prec`: 8826,
	`prnE`: 10933,
	`prod`: 8719,
	`prop`: 8733,
	`Pscr`: 119979,
	`pscr`: 120005,
	`qint`: 10764,
	`qopf`: 120162,
	`Qopf`: 8474,
	`qscr`: 120006,
	`Qscr`: 119980,
	`quot`: 34,
	`QUOT`: 34,
	`Rang`: 10219,
	`rang`: 10217,
	`rArr`: 8658,
	`Rarr`: 8608,
	`rarr`: 8594,
	`rcub`: 125,
	`rdca`: 10551,
	`rdsh`: 8627,
	`real`: 8476,
	`rect`: 9645,
	`rHar`: 10596,
	`rhov`: 1009,
	`ring`: 730,
	`ropf`: 120163,
	`Ropf`: 8477,
	`rpar`: 41,
	`rscr`: 120007,
	`Rscr`: 8475,
	`rsqb`: 93,
	`rtri`: 9657,
	`scap`: 10936,
	`scnE`: 10934,
	`sdot`: 8901,
	`sect`: 167,
	`semi`: 59,
	`sext`: 10038,
	`shcy`: 1096,
	`SHcy`: 1064,
	`sime`: 8771,
	`simg`: 10910,
	`siml`: 10909,
	`smid`: 8739,
	`smte`: 10924,
	`solb`: 10692,
	`Sopf`: 120138,
	`sopf`: 120164,
	`spar`: 8741,
	`Sqrt`: 8730,
	`squf`: 9642,
	`sscr`: 120008,
	`Sscr`: 119982,
	`Star`: 8902,
	`star`: 9734,
	`subE`: 10949,
	`sube`: 8838,
	`succ`: 8827,
	`sung`: 9834,
	`sup1`: 185,
	`sup2`: 178,
	`sup3`: 179,
	`supe`: 8839,
	`supE`: 10950,
	`tbrk`: 9140,
	`tdot`: 8411,
	`tint`: 8749,
	`toea`: 10536,
	`Topf`: 120139,
	`topf`: 120165,
	`tosa`: 10537,
	`trie`: 8796,
	`Tscr`: 119983,
	`tscr`: 120009,
	`tscy`: 1094,
	`TScy`: 1062,
	`uarr`: 8593,
	`Uarr`: 8607,
	`uArr`: 8657,
	`uHar`: 10595,
	`Uopf`: 120140,
	`uopf`: 120166,
	`Upsi`: 978,
	`upsi`: 965,
	`Uscr`: 119984,
	`uscr`: 120010,
	`utri`: 9653,
	`uuml`: 252,
	`Uuml`: 220,
	`vArr`: 8661,
	`varr`: 8597,
	`Vbar`: 10987,
	`vBar`: 10984,
	`vert`: 124,
	`Vert`: 8214,
	`vopf`: 120167,
	`Vopf`: 120141,
	`Vscr`: 119985,
	`vscr`: 120011,
	`Wopf`: 120142,
	`wopf`: 120168,
	`Wscr`: 119986,
	`wscr`: 120012,
	`xcap`: 8898,
	`xcup`: 8899,
	`xmap`: 10236,
	`xnis`: 8955,
	`xopf`: 120169,
	`Xopf`: 120143,
	`xscr`: 120013,
	`Xscr`: 119987,
	`xvee`: 8897,
	`YAcy`: 1071,
	`yacy`: 1103,
	`YIcy`: 1031,
	`yicy`: 1111,
	`Yopf`: 120144,
	`yopf`: 120170,
	`yscr`: 120014,
	`Yscr`: 119988,
	`yucy`: 1102,
	`YUcy`: 1070,
	`yuml`: 255,
	`Yuml`: 376,
	`Zdot`: 379,
	`zdot`: 380,
	`zeta`: 950,
	`Zeta`: 918,
	`ZHcy`: 1046,
	`zhcy`: 1078,
	`zopf`: 120171,
	`Zopf`: 8484,
	`Zscr`: 119989,
	`zscr`: 120015,
	`zwnj`: 8204,

	`Acirc`: 194,
	`acirc`: 226,
	`acute`: 180,
	`AElig`: 198,
	`aelig`: 230,
	`aleph`: 8501,
	`alpha`: 945,
	`Alpha`: 913,
	`Amacr`: 256,
	`amacr`: 257,
	`amalg`: 10815,
	`angle`: 8736,
	`angrt`: 8735,
	`angst`: 197,
	`Aogon`: 260,
	`aogon`: 261,
	`aring`: 229,
	`Aring`: 197,
	`asymp`: 8776,
	`awint`: 10769,
	`bcong`: 8780,
	`bdquo`: 8222,
	`bepsi`: 1014,
	`blank`: 9251,
	`blk12`: 9618,
	`blk14`: 9617,
	`blk34`: 9619,
	`block`: 9608,
	`boxdl`: 9488,
	`boxdL`: 9557,
	`boxDl`: 9558,
	`boxDL`: 9559,
	`boxdr`: 9484,
	`boxDr`: 9555,
	`boxDR`: 9556,
	`boxdR`: 9554,
	`boxHD`: 9574,
	`boxHd`: 9572,
	`boxhd`: 9516,
	`boxhD`: 9573,
	`boxhU`: 9576,
	`boxHu`: 9575,
	`boxhu`: 9524,
	`boxHU`: 9577,
	`boxuL`: 9563,
	`boxul`: 9496,
	`boxUL`: 9565,
	`boxUl`: 9564,
	`boxUR`: 9562,
	`boxur`: 9492,
	`boxuR`: 9560,
	`boxUr`: 9561,
	`boxVh`: 9579,
	`boxvH`: 9578,
	`boxVH`: 9580,
	`boxvh`: 9532,
	`boxvl`: 9508,
	`boxVl`: 9570,
	`boxvL`: 9569,
	`boxVL`: 9571,
	`boxvr`: 9500,
	`boxVR`: 9568,
	`boxvR`: 9566,
	`boxVr`: 9567,
	`Breve`: 728,
	`breve`: 728,
	`bsemi`: 8271,
	`bsime`: 8909,
	`bsolb`: 10693,
	`bumpE`: 10926,
	`bumpe`: 8783,
	`caret`: 8257,
	`caron`: 711,
	`ccaps`: 10829,
	`Ccirc`: 264,
	`ccirc`: 265,
	`ccups`: 10828,
	`cedil`: 184,
	`check`: 10003,
	`clubs`: 9827,
	`colon`: 58,
	`Colon`: 8759,
	`comma`: 44,
	`crarr`: 8629,
	`cross`: 10007,
	`Cross`: 10799,
	`csube`: 10961,
	`csupe`: 10962,
	`ctdot`: 8943,
	`cuepr`: 8926,
	`cuesc`: 8927,
	`cupor`: 10821,
	`cuvee`: 8910,
	`cuwed`: 8911,
	`cwint`: 8753,
	`dashv`: 8867,
	`Dashv`: 10980,
	`dblac`: 733,
	`ddarr`: 8650,
	`Delta`: 916,
	`delta`: 948,
	`dharl`: 8643,
	`dharr`: 8642,
	`diams`: 9830,
	`disin`: 8946,
	`doteq`: 8784,
	`dtdot`: 8945,
	`dtrif`: 9662,
	`duarr`: 8693,
	`duhar`: 10607,
	`Ecirc`: 202,
	`ecirc`: 234,
	`eDDot`: 10871,
	`efDot`: 8786,
	`Emacr`: 274,
	`emacr`: 275,
	`empty`: 8709,
	`eogon`: 281,
	`Eogon`: 280,
	`eplus`: 10865,
	`epsiv`: 1013,
	`eqsim`: 8770,
	`Equal`: 10869,
	`equiv`: 8801,
	`erarr`: 10609,
	`erDot`: 8787,
	`esdot`: 8784,
	`exist`: 8707,
	`fflig`: 64256,
	`filig`: 64257,
	`fllig`: 64258,
	`fltns`: 9649,
	`forkv`: 10969,
	`frasl`: 8260,
	`frown`: 8994,
	`gamma`: 947,
	`Gamma`: 915,
	`Gcirc`: 284,
	`gcirc`: 285,
	`gescc`: 10921,
	`gimel`: 8503,
	`gneqq`: 8809,
	`gnsim`: 8935,
	`grave`: 96,
	`gsime`: 10894,
	`gsiml`: 10896,
	`gtcir`: 10874,
	`gtdot`: 8919,
	`Hacek`: 711,
	`harrw`: 8621,
	`Hcirc`: 292,
	`hcirc`: 293,
	`hoarr`: 8703,
	`icirc`: 238,
	`Icirc`: 206,
	`iexcl`: 161,
	`iiint`: 8749,
	`iiota`: 8489,
	`ijlig`: 307,
	`IJlig`: 306,
	`Imacr`: 298,
	`imacr`: 299,
	`image`: 8465,
	`imath`: 305,
	`imped`: 437,
	`infin`: 8734,
	`Iogon`: 302,
	`iogon`: 303,
	`iprod`: 10812,
	`isinE`: 8953,
	`isins`: 8948,
	`isinv`: 8712,
	`Iukcy`: 1030,
	`iukcy`: 1110,
	`jcirc`: 309,
	`Jcirc`: 308,
	`jmath`: 567,
	`Jukcy`: 1028,
	`jukcy`: 1108,
	`kappa`: 954,
	`Kappa`: 922,
	`lAarr`: 8666,
	`langd`: 10641,
	`laquo`: 171,
	`larrb`: 8676,
	`lBarr`: 10510,
	`lbarr`: 10508,
	`lbbrk`: 10098,
	`lbrke`: 10635,
	`lceil`: 8968,
	`ldquo`: 8220,
	`lescc`: 10920,
	`lhard`: 8637,
	`lharu`: 8636,
	`lhblk`: 9604,
	`llarr`: 8647,
	`lltri`: 9722,
	`lneqq`: 8808,
	`lnsim`: 8934,
	`loang`: 10220,
	`loarr`: 8701,
	`lobrk`: 10214,
	`lopar`: 10629,
	`lrarr`: 8646,
	`lrhar`: 8651,
	`lrtri`: 8895,
	`lsime`: 10893,
	`lsimg`: 10895,
	`lsquo`: 8216,
	`ltcir`: 10873,
	`ltdot`: 8918,
	`ltrie`: 8884,
	`ltrif`: 9666,
	`mdash`: 8212,
	`mDDot`: 8762,
	`micro`: 181,
	`minus`: 8722,
	`mumap`: 8888,
	`nabla`: 8711,
	`napos`: 329,
	`natur`: 9838,
	`ncong`: 8775,
	`ndash`: 8211,
	`neArr`: 8663,
	`nearr`: 8599,
	`ngsim`: 8821,
	`nharr`: 8622,
	`nhArr`: 8654,
	`nhpar`: 10994,
	`nlarr`: 8602,
	`nlArr`: 8653,
	`nless`: 8814,
	`nlsim`: 8820,
	`nltri`: 8938,
	`notin`: 8713,
	`notni`: 8716,
	`nprec`: 8832,
	`nrarr`: 8603,
	`nrArr`: 8655,
	`nrtri`: 8939,
	`nsime`: 8772,
	`nsmid`: 8740,
	`nspar`: 8742,
	`nsube`: 8840,
	`nsucc`: 8833,
	`nsupe`: 8841,
	`numsp`: 8199,
	`nwarr`: 8598,
	`nwArr`: 8662,
	`ocirc`: 244,
	`Ocirc`: 212,
	`odash`: 8861,
	`oelig`: 339,
	`OElig`: 338,
	`ofcir`: 10687,
	`ohbar`: 10677,
	`olarr`: 8634,
	`olcir`: 10686,
	`oline`: 8254,
	`Omacr`: 332,
	`omacr`: 333,
	`omega`: 969,
	`Omega`: 937,
	`operp`: 10681,
	`oplus`: 8853,
	`orarr`: 8635,
	`order`: 8500,
	`ovbar`: 9021,
	`parsl`: 11005,
	`phone`: 9742,
	`plusb`: 8862,
	`pluse`: 10866,
	`pound`: 163,
	`prcue`: 8828,
	`Prime`: 8243,
	`prime`: 8242,
	`prnap`: 10937,
	`prsim`: 8830,
	`quest`: 63,
	`rAarr`: 8667,
	`radic`: 8730,
	`rangd`: 10642,
	`range`: 10661,
	`raquo`: 187,
	`rarrb`: 8677,
	`rarrc`: 10547,
	`rarrw`: 8605,
	`ratio`: 8758,
	`rBarr`: 10511,
	`rbarr`: 10509,
	`RBarr`: 10512,
	`rbbrk`: 10099,
	`rbrke`: 10636,
	`rceil`: 8969,
	`rdquo`: 8221,
	`reals`: 8477,
	`rhard`: 8641,
	`rharu`: 8640,
	`rlarr`: 8644,
	`rlhar`: 8652,
	`rnmid`: 10990,
	`roang`: 10221,
	`roarr`: 8702,
	`robrk`: 10215,
	`ropar`: 10630,
	`rrarr`: 8649,
	`rsquo`: 8217,
	`rtrie`: 8885,
	`rtrif`: 9656,
	`sbquo`: 8218,
	`sccue`: 8829,
	`scirc`: 349,
	`Scirc`: 348,
	`scnap`: 10938,
	`scsim`: 8831,
	`sdotb`: 8865,
	`sdote`: 10854,
	`searr`: 8600,
	`seArr`: 8664,
	`setmn`: 8726,
	`sharp`: 9839,
	`sigma`: 963,
	`Sigma`: 931,
	`simeq`: 8771,
	`simgE`: 10912,
	`simlE`: 10911,
	`simne`: 8774,
	`slarr`: 8592,
	`smile`: 8995,
	`sqcap`: 8851,
	`sqcup`: 8852,
	`sqsub`: 8847,
	`sqsup`: 8848,
	`srarr`: 8594,
	`starf`: 9733,
	`strns`: 175,
	`subnE`: 10955,
	`subne`: 8842,
	`supnE`: 10956,
	`supne`: 8843,
	`swArr`: 8665,
	`swarr`: 8601,
	`szlig`: 223,
	`theta`: 952,
	`Theta`: 920,
	`thkap`: 8776,
	`THORN`: 222,
	`thorn`: 254,
	`tilde`: 732,
	`Tilde`: 8764,
	`times`: 215,
	`trade`: 8482,
	`TRADE`: 8482,
	`trisb`: 10701,
	`tshcy`: 1115,
	`TSHcy`: 1035,
	`twixt`: 8812,
	`ubrcy`: 1118,
	`Ubrcy`: 1038,
	`ucirc`: 251,
	`Ucirc`: 219,
	`udarr`: 8645,
	`udhar`: 10606,
	`uharl`: 8639,
	`uharr`: 8638,
	`uhblk`: 9600,
	`ultri`: 9720,
	`umacr`: 363,
	`Umacr`: 362,
	`Union`: 8899,
	`Uogon`: 370,
	`uogon`: 371,
	`uplus`: 8846,
	`upsih`: 978,
	`UpTee`: 8869,
	`uring`: 367,
	`Uring`: 366,
	`urtri`: 9721,
	`utdot`: 8944,
	`utrif`: 9652,
	`uuarr`: 8648,
	`varpi`: 982,
	`vBarv`: 10985,
	`Vdash`: 8873,
	`vDash`: 8872,
	`VDash`: 8875,
	`vdash`: 8866,
	`veeeq`: 8794,
	`vltri`: 8882,
	`vprop`: 8733,
	`vrtri`: 8883,
	`wcirc`: 373,
	`Wcirc`: 372,
	`wedge`: 8743,
	`Wedge`: 8896,
	`xcirc`: 9711,
	`xdtri`: 9661,
	`xhArr`: 10234,
	`xharr`: 10231,
	`xlarr`: 10229,
	`xlArr`: 10232,
	`xodot`: 10752,
	`xrArr`: 10233,
	`xrarr`: 10230,
	`xutri`: 9651,
	`ycirc`: 375,
	`Ycirc`: 374,

	`aacute`: 225,
	`Aacute`: 193,
	`abreve`: 259,
	`Abreve`: 258,
	`agrave`: 224,
	`Agrave`: 192,
	`andand`: 10837,
	`angmsd`: 8737,
	`angsph`: 8738,
	`apacir`: 10863,
	`approx`: 8776,
	`Assign`: 8788,
	`Atilde`: 195,
	`atilde`: 227,
	`barvee`: 8893,
	`Barwed`: 8966,
	`barwed`: 8965,
	`becaus`: 8757,
	`bernou`: 8492,
	`bigcap`: 8898,
	`bigcup`: 8899,
	`bigvee`: 8897,
	`bkarow`: 10509,
	`bottom`: 8869,
	`bowtie`: 8904,
	`boxbox`: 10697,
	`bprime`: 8245,
	`brvbar`: 166,
	`bullet`: 8226,
	`bumpeq`: 8783,
	`Bumpeq`: 8782,
	`cacute`: 263,
	`Cacute`: 262,
	`capand`: 10820,
	`capcap`: 10827,
	`capcup`: 10823,
	`capdot`: 10816,
	`ccaron`: 269,
	`Ccaron`: 268,
	`Ccedil`: 199,
	`ccedil`: 231,
	`circeq`: 8791,
	`cirmid`: 10991,
	`Colone`: 10868,
	`colone`: 8788,
	`commat`: 64,
	`compfn`: 8728,
	`conint`: 8750,
	`Conint`: 8751,
	`coprod`: 8720,
	`copysr`: 8471,
	`cularr`: 8630,
	`cupcap`: 10822,
	`CupCap`: 8781,
	`cupcup`: 10826,
	`cupdot`: 8845,
	`curarr`: 8631,
	`curren`: 164,
	`cylcty`: 9005,
	`Dagger`: 8225,
	`dagger`: 8224,
	`daleth`: 8504,
	`Dcaron`: 270,
	`dcaron`: 271,
	`dfisht`: 10623,
	`divide`: 247,
	`divonx`: 8903,
	`dlcorn`: 8990,
	`dlcrop`: 8973,
	`dollar`: 36,
	`DotDot`: 8412,
	`drcorn`: 8991,
	`drcrop`: 8972,
	`Dstrok`: 272,
	`dstrok`: 273,
	`Eacute`: 201,
	`eacute`: 233,
	`easter`: 10862,
	`ecaron`: 283,
	`Ecaron`: 282,
	`ecolon`: 8789,
	`Egrave`: 200,
	`egrave`: 232,
	`egsdot`: 10904,
	`elsdot`: 10903,
	`emptyv`: 8709,
	`emsp13`: 8196,
	`emsp14`: 8197,
	`eparsl`: 10723,
	`eqcirc`: 8790,
	`equals`: 61,
	`equest`: 8799,
	`Exists`: 8707,
	`female`: 9792,
	`ffilig`: 64259,
	`ffllig`: 64260,
	`forall`: 8704,
	`ForAll`: 8704,
	`frac12`: 189,
	`frac13`: 8531,
	`frac14`: 188,
	`frac15`: 8533,
	`frac16`: 8537,
	`frac18`: 8539,
	`frac23`: 8532,
	`frac25`: 8534,
	`frac34`: 190,
	`frac35`: 8535,
	`frac38`: 8540,
	`frac45`: 8536,
	`frac56`: 8538,
	`frac58`: 8541,
	`frac78`: 8542,
	`gacute`: 501,
	`Gammad`: 988,
	`gammad`: 989,
	`gbreve`: 287,
	`Gbreve`: 286,
	`Gcedil`: 290,
	`gesdot`: 10880,
	`gesles`: 10900,
	`gtlPar`: 10645,
	`gtrarr`: 10616,
	`gtrdot`: 8919,
	`gtrsim`: 8819,
	`hairsp`: 8202,
	`hamilt`: 8459,
	`HARDcy`: 1066,
	`hardcy`: 1098,
	`hearts`: 9829,
	`hellip`: 8230,
	`hercon`: 8889,
	`homtht`: 8763,
	`horbar`: 8213,
	`hslash`: 8463,
	`hstrok`: 295,
	`Hstrok`: 294,
	`hybull`: 8259,
	`hyphen`: 8208,
	`Iacute`: 205,
	`iacute`: 237,
	`Igrave`: 204,
	`igrave`: 236,
	`iiiint`: 10764,
	`iinfin`: 10716,
	`incare`: 8453,
	`inodot`: 305,
	`intcal`: 8890,
	`iquest`: 191,
	`isinsv`: 8947,
	`itilde`: 297,
	`Itilde`: 296,
	`Jsercy`: 1032,
	`jsercy`: 1112,
	`kappav`: 1008,
	`kcedil`: 311,
	`Kcedil`: 310,
	`kgreen`: 312,
	`Lacute`: 313,
	`lacute`: 314,
	`lagran`: 8466,
	`lambda`: 955,
	`Lambda`: 923,
	`langle`: 10216,
	`larrfs`: 10525,
	`larrhk`: 8617,
	`larrlp`: 8619,
	`larrpl`: 10553,
	`larrtl`: 8610,
	`lAtail`: 10523,
	`latail`: 10521,
	`lbrace`: 123,
	`lbrack`: 91,
	`Lcaron`: 317,
	`lcaron`: 318,
	`lcedil`: 316,
	`Lcedil`: 315,
	`ldquor`: 8222,
	`lesdot`: 10879,
	`lesges`: 10899,
	`lfisht`: 10620,
	`lfloor`: 8970,
	`lharul`: 10602,
	`llhard`: 10603,
	`Lmidot`: 319,
	`lmidot`: 320,
	`lmoust`: 9136,
	`loplus`: 10797,
	`lowast`: 8727,
	`lowbar`: 95,
	`lparlt`: 10643,
	`lrhard`: 10605,
	`lsaquo`: 8249,
	`lsquor`: 8218,
	`Lstrok`: 321,
	`lstrok`: 322,
	`lthree`: 8907,
	`ltimes`: 8905,
	`ltlarr`: 10614,
	`ltrPar`: 10646,
	`mapsto`: 8614,
	`marker`: 9646,
	`mcomma`: 10793,
	`midast`: 42,
	`midcir`: 10992,
	`middot`: 183,
	`minusb`: 8863,
	`minusd`: 8760,
	`mnplus`: 8723,
	`models`: 8871,
	`mstpos`: 8766,
	`Nacute`: 323,
	`nacute`: 324,
	`ncaron`: 328,
	`Ncaron`: 327,
	`Ncedil`: 325,
	`ncedil`: 326,
	`nearhk`: 10532,
	`nequiv`: 8802,
	`nesear`: 10536,
	`nexist`: 8708,
	`nltrie`: 8940,
	`nprcue`: 8928,
	`nrtrie`: 8941,
	`nsccue`: 8929,
	`nsimeq`: 8772,
	`ntilde`: 241,
	`Ntilde`: 209,
	`numero`: 8470,
	`nVDash`: 8879,
	`nvDash`: 8877,
	`nVdash`: 8878,
	`nvdash`: 8876,
	`nvHarr`: 10500,
	`nvlArr`: 10498,
	`nvrArr`: 10499,
	`nwarhk`: 10531,
	`nwnear`: 10535,
	`oacute`: 243,
	`Oacute`: 211,
	`Odblac`: 336,
	`odblac`: 337,
	`odsold`: 10684,
	`Ograve`: 210,
	`ograve`: 242,
	`ominus`: 8854,
	`origof`: 8886,
	`oslash`: 248,
	`Oslash`: 216,
	`otilde`: 245,
	`Otilde`: 213,
	`Otimes`: 10807,
	`otimes`: 8855,
	`parsim`: 10995,
	`percnt`: 37,
	`period`: 46,
	`permil`: 8240,
	`phmmat`: 8499,
	`planck`: 8463,
	`plankv`: 8463,
	`plusdo`: 8724,
	`plusdu`: 10789,
	`plusmn`: 177,
	`preceq`: 10927,
	`primes`: 8473,
	`prnsim`: 8936,
	`propto`: 8733,
	`prurel`: 8880,
	`puncsp`: 8200,
	`qprime`: 8279,
	`racute`: 341,
	`Racute`: 340,
	`rangle`: 10217,
	`rarrap`: 10613,
	`rarrfs`: 10526,
	`rarrhk`: 8618,
	`rarrlp`: 8620,
	`rarrpl`: 10565,
	`rarrtl`: 8611,
	`Rarrtl`: 10518,
	`rAtail`: 10524,
	`ratail`: 10522,
	`rbrace`: 125,
	`rbrack`: 93,
	`Rcaron`: 344,
	`rcaron`: 345,
	`rcedil`: 343,
	`Rcedil`: 342,
	`rdquor`: 8221,
	`rfisht`: 10621,
	`rfloor`: 8971,
	`rharul`: 10604,
	`rmoust`: 9137,
	`roplus`: 10798,
	`rpargt`: 10644,
	`rsaquo`: 8250,
	`rsquor`: 8217,
	`rthree`: 8908,
	`rtimes`: 8906,
	`Sacute`: 346,
	`sacute`: 347,
	`scaron`: 353,
	`Scaron`: 352,
	`Scedil`: 350,
	`scedil`: 351,
	`scnsim`: 8937,
	`searhk`: 10533,
	`seswar`: 10537,
	`sfrown`: 8994,
	`SHCHcy`: 1065,
	`shchcy`: 1097,
	`sigmaf`: 962,
	`sigmav`: 962,
	`simdot`: 10858,
	`smashp`: 10803,
	`SOFTcy`: 1068,
	`softcy`: 1100,
	`solbar`: 9023,
	`spades`: 9824,
	`sqsube`: 8849,
	`sqsupe`: 8850,
	`square`: 9633,
	`Square`: 9633,
	`squarf`: 9642,
	`ssetmn`: 8726,
	`ssmile`: 8995,
	`sstarf`: 8902,
	`subdot`: 10941,
	`subset`: 8834,
	`Subset`: 8912,
	`subsim`: 10951,
	`subsub`: 10965,
	`subsup`: 10963,
	`succeq`: 10928,
	`supdot`: 10942,
	`supset`: 8835,
	`Supset`: 8913,
	`supsim`: 10952,
	`supsub`: 10964,
	`supsup`: 10966,
	`swarhk`: 10534,
	`swnwar`: 10538,
	`target`: 8982,
	`tcaron`: 357,
	`Tcaron`: 356,
	`tcedil`: 355,
	`Tcedil`: 354,
	`telrec`: 8981,
	`there4`: 8756,
	`thetav`: 977,
	`thinsp`: 8201,
	`thksim`: 8764,
	`timesb`: 8864,
	`timesd`: 10800,
	`topbot`: 9014,
	`topcir`: 10993,
	`tprime`: 8244,
	`tridot`: 9708,
	`Tstrok`: 358,
	`tstrok`: 359,
	`Uacute`: 218,
	`uacute`: 250,
	`ubreve`: 365,
	`Ubreve`: 364,
	`Udblac`: 368,
	`udblac`: 369,
	`ufisht`: 10622,
	`Ugrave`: 217,
	`ugrave`: 249,
	`ulcorn`: 8988,
	`ulcrop`: 8975,
	`urcorn`: 8989,
	`urcrop`: 8974,
	`Utilde`: 360,
	`utilde`: 361,
	`vangrt`: 10652,
	`varphi`: 981,
	`varrho`: 1009,
	`Vdashl`: 10982,
	`veebar`: 8891,
	`vellip`: 8942,
	`Verbar`: 8214,
	`verbar`: 124,
	`Vvdash`: 8874,
	`wedbar`: 10847,
	`wedgeq`: 8793,
	`weierp`: 8472,
	`wreath`: 8768,
	`xoplus`: 10753,
	`xotime`: 10754,
	`xsqcup`: 10758,
	`xuplus`: 10756,
	`xwedge`: 8896,
	`yacute`: 253,
	`Yacute`: 221,
	`Zacute`: 377,
	`zacute`: 378,
	`zcaron`: 382,
	`Zcaron`: 381,
	`zeetrf`: 8488,

	`alefsym`: 8501,
	`angrtvb`: 8894,
	`angzarr`: 9084,
	`asympeq`: 8781,
	`backsim`: 8765,
	`because`: 8757,
	`Because`: 8757,
	`bemptyv`: 10672,
	`between`: 8812,
	`bigcirc`: 9711,
	`bigodot`: 10752,
	`bigstar`: 9733,
	`boxplus`: 8862,
	`Cayleys`: 8493,
	`Cconint`: 8752,
	`ccupssm`: 10832,
	`Cedilla`: 184,
	`cemptyv`: 10674,
	`cirscir`: 10690,
	`coloneq`: 8788,
	`congdot`: 10861,
	`cudarrl`: 10552,
	`cudarrr`: 10549,
	`cularrp`: 10557,
	`curarrm`: 10556,
	`dbkarow`: 10511,
	`ddagger`: 8225,
	`ddotseq`: 10871,
	`demptyv`: 10673,
	`diamond`: 8900,
	`Diamond`: 8900,
	`digamma`: 989,
	`dotplus`: 8724,
	`DownTee`: 8868,
	`dwangle`: 10662,
	`Element`: 8712,
	`epsilon`: 949,
	`Epsilon`: 917,
	`eqcolon`: 8789,
	`equivDD`: 10872,
	`gesdoto`: 10882,
	`gtquest`: 10876,
	`gtrless`: 8823,
	`harrcir`: 10568,
	`Implies`: 8658,
	`intprod`: 10812,
	`isindot`: 8949,
	`larrbfs`: 10527,
	`larrsim`: 10611,
	`lbrksld`: 10639,
	`lbrkslu`: 10637,
	`ldrdhar`: 10599,
	`LeftTee`: 8867,
	`lesdoto`: 10881,
	`lessdot`: 8918,
	`lessgtr`: 8822,
	`lesssim`: 8818,
	`lotimes`: 10804,
	`lozenge`: 9674,
	`ltquest`: 10875,
	`luruhar`: 10598,
	`maltese`: 10016,
	`minusdu`: 10794,
	`napprox`: 8777,
	`natural`: 9838,
	`nearrow`: 8599,
	`NewLine`: 10,
	`nexists`: 8708,
	`NoBreak`: 8288,
	`notinva`: 8713,
	`notinvb`: 8951,
	`notinvc`: 8950,
	`NotLess`: 8814,
	`notniva`: 8716,
	`notnivb`: 8958,
	`notnivc`: 8957,
	`npolint`: 10772,
	`nsqsube`: 8930,
	`nsqsupe`: 8931,
	`nvinfin`: 10718,
	`nwarrow`: 8598,
	`olcross`: 10683,
	`omicron`: 959,
	`Omicron`: 927,
	`orderof`: 8500,
	`orslope`: 10839,
	`OverBar`: 8254,
	`pertenk`: 8241,
	`planckh`: 8462,
	`pluscir`: 10786,
	`plussim`: 10790,
	`plustwo`: 10791,
	`precsim`: 8830,
	`Product`: 8719,
	`quatint`: 10774,
	`questeq`: 8799,
	`rarrbfs`: 10528,
	`rarrsim`: 10612,
	`rbrksld`: 10638,
	`rbrkslu`: 10640,
	`rdldhar`: 10601,
	`realine`: 8475,
	`rotimes`: 10805,
	`ruluhar`: 10600,
	`searrow`: 8600,
	`simplus`: 10788,
	`simrarr`: 10610,
	`subedot`: 10947,
	`submult`: 10945,
	`subplus`: 10943,
	`subrarr`: 10617,
	`succsim`: 8831,
	`supdsub`: 10968,
	`supedot`: 10948,
	`suphsol`: 10185,
	`suphsub`: 10967,
	`suplarr`: 10619,
	`supmult`: 10946,
	`supplus`: 10944,
	`swarrow`: 8601,
	`topfork`: 10970,
	`triplus`: 10809,
	`tritime`: 10811,
	`Uparrow`: 8657,
	`UpArrow`: 8593,
	`uparrow`: 8593,
	`Upsilon`: 933,
	`upsilon`: 965,
	`uwangle`: 10663,
	`vzigzag`: 10650,
	`zigrarr`: 8669,

	`andslope`: 10840,
	`angmsdaa`: 10664,
	`angmsdab`: 10665,
	`angmsdac`: 10666,
	`angmsdad`: 10667,
	`angmsdae`: 10668,
	`angmsdaf`: 10669,
	`angmsdag`: 10670,
	`angmsdah`: 10671,
	`angrtvbd`: 10653,
	`approxeq`: 8778,
	`awconint`: 8755,
	`backcong`: 8780,
	`barwedge`: 8965,
	`bbrktbrk`: 9142,
	`bigoplus`: 10753,
	`bigsqcup`: 10758,
	`biguplus`: 10756,
	`bigwedge`: 8896,
	`boxminus`: 8863,
	`boxtimes`: 8864,
	`bsolhsub`: 10184,
	`capbrcup`: 10825,
	`circledR`: 174,
	`circledS`: 9416,
	`cirfnint`: 10768,
	`clubsuit`: 9827,
	`cupbrcap`: 10824,
	`curlyvee`: 8910,
	`cwconint`: 8754,
	`DDotrahd`: 10513,
	`doteqdot`: 8785,
	`DotEqual`: 8784,
	`dotminus`: 8760,
	`drbkarow`: 10512,
	`dzigrarr`: 10239,
	`elinters`: 9191,
	`emptyset`: 8709,
	`eqvparsl`: 10725,
	`fpartint`: 10765,
	`geqslant`: 10878,
	`gesdotol`: 10884,
	`gnapprox`: 10890,
	`hksearow`: 10533,
	`hkswarow`: 10534,
	`imagline`: 8464,
	`imagpart`: 8465,
	`infintie`: 10717,
	`integers`: 8484,
	`Integral`: 8747,
	`intercal`: 8890,
	`intlarhk`: 10775,
	`laemptyv`: 10676,
	`ldrushar`: 10571,
	`leqslant`: 10877,
	`lesdotor`: 10883,
	`LessLess`: 10913,
	`llcorner`: 8990,
	`lnapprox`: 10889,
	`lrcorner`: 8991,
	`lurdshar`: 10570,
	`mapstoup`: 8613,
	`multimap`: 8888,
	`naturals`: 8469,
	`NotEqual`: 8800,
	`NotTilde`: 8769,
	`otimesas`: 10806,
	`parallel`: 8741,
	`PartialD`: 8706,
	`plusacir`: 10787,
	`pointint`: 10773,
	`Precedes`: 8826,
	`precneqq`: 10933,
	`precnsim`: 8936,
	`profalar`: 9006,
	`profline`: 8978,
	`profsurf`: 8979,
	`raemptyv`: 10675,
	`realpart`: 8476,
	`RightTee`: 8866,
	`rppolint`: 10770,
	`rtriltri`: 10702,
	`scpolint`: 10771,
	`setminus`: 8726,
	`shortmid`: 8739,
	`smeparsl`: 10724,
	`sqsubset`: 8847,
	`sqsupset`: 8848,
	`subseteq`: 8838,
	`Succeeds`: 8827,
	`succneqq`: 10934,
	`succnsim`: 8937,
	`SuchThat`: 8715,
	`Superset`: 8835,
	`supseteq`: 8839,
	`thetasym`: 977,
	`thicksim`: 8764,
	`timesbar`: 10801,
	`triangle`: 9653,
	`triminus`: 10810,
	`trpezium`: 9186,
	`Uarrocir`: 10569,
	`ulcorner`: 8988,
	`UnderBar`: 95,
	`urcorner`: 8989,
	`varkappa`: 1008,
	`varsigma`: 962,
	`vartheta`: 977,

	`backprime`: 8245,
	`backsimeq`: 8909,
	`Backslash`: 8726,
	`bigotimes`: 10754,
	`centerdot`: 183,
	`CenterDot`: 183,
	`checkmark`: 10003,
	`CircleDot`: 8857,
	`complexes`: 8450,
	`Congruent`: 8801,
	`Coproduct`: 8720,
	`dotsquare`: 8865,
	`DoubleDot`: 168,
	`downarrow`: 8595,
	`Downarrow`: 8659,
	`DownArrow`: 8595,
	`DownBreve`: 785,
	`gtrapprox`: 10886,
	`gtreqless`: 8923,
	`heartsuit`: 9829,
	`HumpEqual`: 8783,
	`leftarrow`: 8592,
	`Leftarrow`: 8656,
	`LeftArrow`: 8592,
	`LeftFloor`: 8970,
	`lesseqgtr`: 8922,
	`LessTilde`: 8818,
	`Mellintrf`: 8499,
	`MinusPlus`: 8723,
	`NotCupCap`: 8813,
	`NotExists`: 8708,
	`nparallel`: 8742,
	`nshortmid`: 8740,
	`nsubseteq`: 8840,
	`nsupseteq`: 8841,
	`OverBrace`: 9182,
	`pitchfork`: 8916,
	`PlusMinus`: 177,
	`rationals`: 8474,
	`spadesuit`: 9824,
	`subseteqq`: 10949,
	`subsetneq`: 8842,
	`supseteqq`: 10950,
	`supsetneq`: 8843,
	`therefore`: 8756,
	`Therefore`: 8756,
	`ThinSpace`: 8201,
	`triangleq`: 8796,
	`TripleDot`: 8411,
	`UnionPlus`: 8846,
	`varpropto`: 8733,

	`Bernoullis`: 8492,
	`circledast`: 8859,
	`CirclePlus`: 8853,
	`complement`: 8705,
	`curlywedge`: 8911,
	`eqslantgtr`: 10902,
	`EqualTilde`: 8770,
	`Fouriertrf`: 8497,
	`gtreqqless`: 10892,
	`ImaginaryI`: 8520,
	`Laplacetrf`: 8466,
	`LeftVector`: 8636,
	`lessapprox`: 10885,
	`lesseqqgtr`: 10891,
	`Lleftarrow`: 8666,
	`lmoustache`: 9136,
	`longmapsto`: 10236,
	`mapstodown`: 8615,
	`mapstoleft`: 8612,
	`nleftarrow`: 8602,
	`nLeftarrow`: 8653,
	`NotElement`: 8713,
	`NotGreater`: 8815,
	`precapprox`: 10935,
	`Proportion`: 8759,
	`rightarrow`: 8594,
	`RightArrow`: 8594,
	`Rightarrow`: 8658,
	`RightFloor`: 8971,
	`rmoustache`: 9137,
	`sqsubseteq`: 8849,
	`sqsupseteq`: 8850,
	`subsetneqq`: 10955,
	`succapprox`: 10936,
	`supsetneqq`: 10956,
	`TildeEqual`: 8771,
	`TildeTilde`: 8776,
	`UnderBrace`: 9183,
	`UpArrowBar`: 10514,
	`UpTeeArrow`: 8613,
	`upuparrows`: 8648,
	`varepsilon`: 1013,
	`varnothing`: 8709,

	`backepsilon`: 1014,
	`blacksquare`: 9642,
	`circledcirc`: 8858,
	`circleddash`: 8861,
	`CircleMinus`: 8854,
	`CircleTimes`: 8855,
	`curlyeqprec`: 8926,
	`curlyeqsucc`: 8927,
	`diamondsuit`: 9830,
	`eqslantless`: 10901,
	`Equilibrium`: 8652,
	`expectation`: 8496,
	`GreaterLess`: 8823,
	`LeftCeiling`: 8968,
	`LessGreater`: 8822,
	`MediumSpace`: 8287,
	`NotPrecedes`: 8832,
	`NotSucceeds`: 8833,
	`nRightarrow`: 8655,
	`nrightarrow`: 8603,
	`OverBracket`: 9140,
	`preccurlyeq`: 8828,
	`precnapprox`: 10937,
	`quaternions`: 8461,
	`RightVector`: 8640,
	`Rrightarrow`: 8667,
	`RuleDelayed`: 10740,
	`SmallCircle`: 8728,
	`SquareUnion`: 8852,
	`straightphi`: 981,
	`SubsetEqual`: 8838,
	`succcurlyeq`: 8829,
	`succnapprox`: 10938,
	`thickapprox`: 8776,
	`updownarrow`: 8597,
	`Updownarrow`: 8661,
	`UpDownArrow`: 8597,
	`VerticalBar`: 8739,

	`blacklozenge`: 10731,
	`DownArrowBar`: 10515,
	`DownTeeArrow`: 8615,
	`ExponentialE`: 8519,
	`exponentiale`: 8519,
	`GreaterEqual`: 8805,
	`GreaterTilde`: 8819,
	`HilbertSpace`: 8459,
	`HumpDownHump`: 8782,
	`Intersection`: 8898,
	`LeftArrowBar`: 8676,
	`LeftTeeArrow`: 8612,
	`LeftTriangle`: 8882,
	`LeftUpVector`: 8639,
	`NotCongruent`: 8802,
	`NotLessEqual`: 8816,
	`NotLessTilde`: 8820,
	`Proportional`: 8733,
	`RightCeiling`: 8969,
	`risingdotseq`: 8787,
	`RoundImplies`: 10608,
	`ShortUpArrow`: 8593,
	`SquareSubset`: 8847,
	`triangledown`: 9663,
	`triangleleft`: 9667,
	`UnderBracket`: 9141,
	`VerticalLine`: 124,

	`ApplyFunction`: 8289,
	`bigtriangleup`: 9651,
	`blacktriangle`: 9652,
	`DifferentialD`: 8518,
	`divideontimes`: 8903,
	`DoubleLeftTee`: 10980,
	`DoubleUpArrow`: 8657,
	`fallingdotseq`: 8786,
	`hookleftarrow`: 8617,
	`leftarrowtail`: 8610,
	`leftharpoonup`: 8636,
	`LeftTeeVector`: 10586,
	`LeftVectorBar`: 10578,
	`LessFullEqual`: 8806,
	`longleftarrow`: 10229,
	`LongLeftArrow`: 10229,
	`Longleftarrow`: 10232,
	`looparrowleft`: 8619,
	`measuredangle`: 8737,
	`NotTildeEqual`: 8772,
	`NotTildeTilde`: 8777,
	`ntriangleleft`: 8938,
	`Poincareplane`: 8460,
	`PrecedesEqual`: 10927,
	`PrecedesTilde`: 8830,
	`RightArrowBar`: 8677,
	`RightTeeArrow`: 8614,
	`RightTriangle`: 8883,
	`RightUpVector`: 8638,
	`shortparallel`: 8741,
	`smallsetminus`: 8726,
	`SucceedsEqual`: 10928,
	`SucceedsTilde`: 8831,
	`SupersetEqual`: 8839,
	`triangleright`: 9657,
	`UpEquilibrium`: 10606,
	`upharpoonleft`: 8639,
	`VerticalTilde`: 8768,
	`VeryThinSpace`: 8202,

	`curvearrowleft`: 8630,
	`DiacriticalDot`: 729,
	`doublebarwedge`: 8966,
	`DoubleRightTee`: 8872,
	`downdownarrows`: 8650,
	`DownLeftVector`: 8637,
	`GreaterGreater`: 10914,
	`hookrightarrow`: 8618,
	`HorizontalLine`: 9472,
	`InvisibleComma`: 8291,
	`InvisibleTimes`: 8290,
	`LeftDownVector`: 8643,
	`leftleftarrows`: 8647,
	`Leftrightarrow`: 8660,
	`LeftRightArrow`: 8596,
	`leftrightarrow`: 8596,
	`leftthreetimes`: 8907,
	`LessSlantEqual`: 10877,
	`LongRightArrow`: 10230,
	`Longrightarrow`: 10233,
	`longrightarrow`: 10230,
	`looparrowright`: 8620,
	`LowerLeftArrow`: 8601,
	`NestedLessLess`: 8810,
	`NotGreaterLess`: 8825,
	`NotLessGreater`: 8824,
	`NotSubsetEqual`: 8840,
	`NotVerticalBar`: 8740,
	`nshortparallel`: 8742,
	`ntriangleright`: 8939,
	`OpenCurlyQuote`: 8216,
	`ReverseElement`: 8715,
	`rightarrowtail`: 8611,
	`rightharpoonup`: 8640,
	`RightTeeVector`: 10587,
	`RightVectorBar`: 10579,
	`ShortDownArrow`: 8595,
	`ShortLeftArrow`: 8592,
	`SquareSuperset`: 8848,
	`TildeFullEqual`: 8773,
	`trianglelefteq`: 8884,
	`upharpoonright`: 8638,
	`UpperLeftArrow`: 8598,
	`ZeroWidthSpace`: 8203,

	`bigtriangledown`: 9661,
	`circlearrowleft`: 8634,
	`CloseCurlyQuote`: 8217,
	`ContourIntegral`: 8750,
	`curvearrowright`: 8631,
	`DoubleDownArrow`: 8659,
	`DoubleLeftArrow`: 8656,
	`downharpoonleft`: 8643,
	`DownRightVector`: 8641,
	`leftharpoondown`: 8637,
	`leftrightarrows`: 8646,
	`LeftRightVector`: 10574,
	`LeftTriangleBar`: 10703,
	`LeftUpTeeVector`: 10592,
	`LeftUpVectorBar`: 10584,
	`LowerRightArrow`: 8600,
	`nleftrightarrow`: 8622,
	`nLeftrightarrow`: 8654,
	`NotGreaterEqual`: 8817,
	`NotGreaterTilde`: 8821,
	`NotLeftTriangle`: 8938,
	`ntrianglelefteq`: 8940,
	`OverParenthesis`: 9180,
	`RightDownVector`: 8642,
	`rightleftarrows`: 8644,
	`rightsquigarrow`: 8605,
	`rightthreetimes`: 8908,
	`ShortRightArrow`: 8594,
	`straightepsilon`: 1013,
	`trianglerighteq`: 8885,
	`UpperRightArrow`: 8599,
	`vartriangleleft`: 8882,

	`circlearrowright`: 8635,
	`DiacriticalAcute`: 180,
	`DiacriticalGrave`: 96,
	`DiacriticalTilde`: 732,
	`DoubleRightArrow`: 8658,
	`DownArrowUpArrow`: 8693,
	`downharpoonright`: 8642,
	`EmptySmallSquare`: 9723,
	`GreaterEqualLess`: 8923,
	`GreaterFullEqual`: 8807,
	`LeftAngleBracket`: 10216,
	`LeftUpDownVector`: 10577,
	`LessEqualGreater`: 8922,
	`NonBreakingSpace`: 160,
	`NotRightTriangle`: 8939,
	`NotSupersetEqual`: 8841,
	`ntrianglerighteq`: 8941,
	`rightharpoondown`: 8641,
	`rightrightarrows`: 8649,
	`RightTriangleBar`: 10704,
	`RightUpTeeVector`: 10588,
	`RightUpVectorBar`: 10580,
	`twoheadleftarrow`: 8606,
	`UnderParenthesis`: 9181,
	`UpArrowDownArrow`: 8645,
	`vartriangleright`: 8883,

	`blacktriangledown`: 9662,
	`blacktriangleleft`: 9666,
	`DoubleUpDownArrow`: 8661,
	`DoubleVerticalBar`: 8741,
	`DownLeftTeeVector`: 10590,
	`DownLeftVectorBar`: 10582,
	`FilledSmallSquare`: 9724,
	`GreaterSlantEqual`: 10878,
	`LeftDoubleBracket`: 10214,
	`LeftDownTeeVector`: 10593,
	`LeftDownVectorBar`: 10585,
	`leftrightharpoons`: 8651,
	`LeftTriangleEqual`: 8884,
	`NegativeThinSpace`: 8203,
	`NotReverseElement`: 8716,
	`NotTildeFullEqual`: 8775,
	`RightAngleBracket`: 10217,
	`rightleftharpoons`: 8652,
	`RightUpDownVector`: 10575,
	`SquareSubsetEqual`: 8849,
	`twoheadrightarrow`: 8608,
	`VerticalSeparator`: 10072,

	`blacktriangleright`: 9656,
	`DownRightTeeVector`: 10591,
	`DownRightVectorBar`: 10583,
	`LongLeftRightArrow`: 10231,
	`Longleftrightarrow`: 10234,
	`longleftrightarrow`: 10231,
	`NegativeThickSpace`: 8203,
	`PrecedesSlantEqual`: 8828,
	`ReverseEquilibrium`: 8651,
	`RightDoubleBracket`: 10215,
	`RightDownTeeVector`: 10589,
	`RightDownVectorBar`: 10581,
	`RightTriangleEqual`: 8885,
	`SquareIntersection`: 8851,
	`SucceedsSlantEqual`: 8829,

	`DoubleLongLeftArrow`: 10232,
	`DownLeftRightVector`: 10576,
	`LeftArrowRightArrow`: 8646,
	`leftrightsquigarrow`: 8621,
	`NegativeMediumSpace`: 8203,
	`RightArrowLeftArrow`: 8644,
	`SquareSupersetEqual`: 8850,

	`CapitalDifferentialD`: 8517,
	`DoubleLeftRightArrow`: 8660,
	`DoubleLongRightArrow`: 10233,
	`EmptyVerySmallSquare`: 9643,
	`NestedGreaterGreater`: 8811,
	`NotDoubleVerticalBar`: 8742,
	`NotLeftTriangleEqual`: 8940,
	`NotSquareSubsetEqual`: 8930,
	`OpenCurlyDoubleQuote`: 8220,
	`ReverseUpEquilibrium`: 10607,

	`CloseCurlyDoubleQuote`: 8221,
	`DoubleContourIntegral`: 8751,
	`FilledVerySmallSquare`: 9642,
	`NegativeVeryThinSpace`: 8203,
	`NotPrecedesSlantEqual`: 8928,
	`NotRightTriangleEqual`: 8941,
	`NotSucceedsSlantEqual`: 8929,

	`DiacriticalDoubleAcute`: 733,
	`NotSquareSupersetEqual`: 8931,

	`ClockwiseContourIntegral`: 8754,
	`DoubleLongLeftRightArrow`: 10234,

	`CounterClockwiseContourIntegral`: 8755,
}

var htmlEntities2 = map[string][2]rune{
	`acE`: {8766, 819},
	`bne`: {61, 8421},
	`ngE`: {8807, 824},
	`nGg`: {8921, 824},
	`nGt`: {8811, 8402},
	`nlE`: {8806, 824},
	`nLl`: {8920, 824},
	`nLt`: {8810, 8402},

	`caps`: {8745, 65024},
	`cups`: {8746, 65024},
	`gesl`: {8923, 65024},
	`gvnE`: {8809, 65024},
	`lesg`: {8922, 65024},
	`lvnE`: {8808, 65024},
	`nang`: {8736, 8402},
	`napE`: {10864, 824},
	`nges`: {10878, 824},
	`nGtv`: {8811, 824},
	`nles`: {10877, 824},
	`nLtv`: {8810, 824},
	`npre`: {10927, 824},
	`nsce`: {10928, 824},
	`nvap`: {8781, 8402},
	`nvge`: {8805, 8402},
	`nvgt`: {62, 8402},
	`nvle`: {8804, 8402},
	`nvlt`: {60, 8402},
	`race`: {8765, 817},

	`fjlig`: {102, 106},
	`lates`: {10925, 65024},
	`napid`: {8779, 824},
	`nbump`: {8782, 824},
	`nedot`: {8784, 824},
	`nesim`: {8770, 824},
	`ngeqq`: {8807, 824},
	`nleqq`: {8806, 824},
	`npart`: {8706, 824},
	`nsubE`: {10949, 824},
	`nsupE`: {10950, 824},
	`nvsim`: {8764, 8402},
	`smtes`: {10924, 65024},
	`vnsub`: {8834, 8402},
	`vnsup`: {8835, 8402},

	`nbumpe`: {8783, 824},
	`notinE`: {8953, 824},
	`nparsl`: {11005, 8421},
	`nrarrc`: {10547, 824},
	`nrarrw`: {8605, 824},
	`sqcaps`: {8851, 65024},
	`sqcups`: {8852, 65024},
	`vsubnE`: {10955, 65024},
	`vsubne`: {8842, 65024},
	`vsupnE`: {10956, 65024},
	`vsupne`: {8843, 65024},

	`bnequiv`: {8801, 8421},
	`npreceq`: {10927, 824},
	`nsubset`: {8834, 8402},
	`nsucceq`: {10928, 824},
	`nsupset`: {8835, 8402},
	`nvltrie`: {8884, 8402},
	`nvrtrie`: {8885, 8402},

	`ncongdot`: {10861, 824},
	`notindot`: {8949, 824},

	`gvertneqq`: {8809, 65024},
	`lvertneqq`: {8808, 65024},
	`ngeqslant`: {10878, 824},
	`nleqslant`: {10877, 824},
	`NotSubset`: {8834, 8402},

	`nsubseteqq`: {10949, 824},
	`nsupseteqq`: {10950, 824},
	`ThickSpace`: {8287, 8202},

	`NotLessLess`: {8810, 824},
	`NotSuperset`: {8835, 8402},

	`NotHumpEqual`: {8783, 824},
	`varsubsetneq`: {8842, 65024},
	`varsupsetneq`: {8843, 65024},

	`NotEqualTilde`: {8770, 824},
	`varsubsetneqq`: {10955, 65024},
	`varsupsetneqq`: {10956, 65024},

	`NotHumpDownHump`: {8782, 824},
	`NotSquareSubset`: {8847, 824},

	`NotPrecedesEqual`: {10927, 824},
	`NotSucceedsEqual`: {10928, 824},
	`NotSucceedsTilde`: {8831, 824},

	`NotGreaterGreater`: {8811, 824},
	`NotLessSlantEqual`: {10877, 824},
	`NotNestedLessLess`: {10913, 824},
	`NotSquareSuperset`: {8848, 824},

	`NotLeftTriangleBar`: {10703, 824},

	`NotGreaterFullEqual`: {8807, 824},
	`NotRightTriangleBar`: {10704, 824},

	`NotGreaterSlantEqual`: {10878, 824},

	`NotNestedGreaterGreater`: {10914, 824},
}
