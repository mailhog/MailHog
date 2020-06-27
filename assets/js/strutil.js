// -*- coding: utf-8 -*-

// GO-MAILHOG: This file borrowed from http://0xcc.net/jsescape/strutil.js

// Utility functions for strings.
//
// Copyright (C) 2007 Satoru Takabayashi <satoru 0xcc.net>
// All rights reserved.  This is free software with ABSOLUTELY NO WARRANTY.
// You can redistribute it and/or modify it under the terms of
// the GNU General Public License version 2.

// NOTES:
//
// Surrogate pairs:
//
//   1st 0xD800 - 0xDBFF (high surrogate)
//   2nd 0xDC00 - 0xDFFF (low surrogate)
//
// UTF-8 sequences:
//
//   0xxxxxxx
//   110xxxxx 10xxxxxx
//   1110xxxx 10xxxxxx 10xxxxxx
//   11110xxx 10xxxxxx 10xxxxxx 10xxxxxx

var EQUAL_SIGN = 0x3D;
var QUESTION_MARK = 0x3F;

// "あい" => [ 0x3042,  0x3044 ]
function convertStringToUnicodeCodePoints(str) {
  var surrogate_1st = 0;
  var unicode_codes = [];
  for (var i = 0; i < str.length; ++i) {
    var utf16_code = str.charCodeAt(i);
    if (surrogate_1st != 0) {
      if (utf16_code >= 0xDC00 && utf16_code <= 0xDFFF) {
        var surrogate_2nd = utf16_code;
        var unicode_code = (surrogate_1st - 0xD800) * (1 << 10) + (1 << 16) +
                           (surrogate_2nd - 0xDC00);
        unicode_codes.push(unicode_code);
      } else {
        // Malformed surrogate pair ignored.
      }
      surrogate_1st = 0;
    } else if (utf16_code >= 0xD800 && utf16_code <= 0xDBFF) {
      surrogate_1st = utf16_code;
    } else {
      unicode_codes.push(utf16_code);
    }
  }
  return unicode_codes;
}

// [ 0xE3, 0x81, 0x82, 0xE3, 0x81, 0x84 ] => [ 0x3042, 0x3044 ]
function convertUtf8BytesToUnicodeCodePoints(utf8_bytes) {
  var unicode_codes = [];
  var unicode_code = 0;
  var num_followed = 0;
  for (var i = 0; i < utf8_bytes.length; ++i) {
    var utf8_byte = utf8_bytes[i];
    if (utf8_byte >= 0x100) {
      // Malformed utf8 byte ignored.
    } else if ((utf8_byte & 0xC0) == 0x80) {
      if (num_followed > 0) {
        unicode_code = (unicode_code << 6) | (utf8_byte & 0x3f);
        num_followed -= 1;
      } else {
        // Malformed UTF-8 sequence ignored.
      }
    } else {
      if (num_followed == 0) {
        unicode_codes.push(unicode_code);
      } else {
        // Malformed UTF-8 sequence ignored.
      }
      if (utf8_byte < 0x80){  // 1-byte
        unicode_code = utf8_byte;
        num_followed = 0;
      } else if ((utf8_byte & 0xE0) == 0xC0) {  // 2-byte
        unicode_code = utf8_byte & 0x1f;
        num_followed = 1;
      } else if ((utf8_byte & 0xF0) == 0xE0) {  // 3-byte
        unicode_code = utf8_byte & 0x0f;
        num_followed = 2;
      } else if ((utf8_byte & 0xF8) == 0xF0) {  // 4-byte
        unicode_code = utf8_byte & 0x07;
        num_followed = 3;
      } else {
        // Malformed UTF-8 sequence ignored.
      }
    }
  }
  if (num_followed == 0) {
    unicode_codes.push(unicode_code);
  } else {
    // Malformed UTF-8 sequence ignored.
  }
  unicode_codes.shift();  // Trim the first element.
  return unicode_codes;
}

// Helper function.
function convertEscapedCodesToCodes(str, prefix, base, num_bits) {
  var parts = str.split(prefix);
  parts.shift();  // Trim the first element.
  var codes = [];
  var max = Math.pow(2, num_bits);
  for (var i = 0; i < parts.length; ++i) {
    var code = parseInt(parts[i], base);
    if (code >= 0 && code < max) {
      codes.push(code);
    } else {
      // Malformed code ignored.
    }
  }
  return codes;
}

// r'\u3042\u3044' => [ 0x3042, 0x3044 ]
// Note that the r '...' notation is borrowed from Python.
function convertEscapedUtf16CodesToUtf16Codes(str) {
  return convertEscapedCodesToCodes(str, "\\u", 16, 16);
}

// r'\U00003042\U00003044' => [ 0x3042, 0x3044 ]
function convertEscapedUtf32CodesToUnicodeCodePoints(str) {
  return convertEscapedCodesToCodes(str, "\\U", 16, 32);
}

// r'\xE3\x81\x82\xE3\x81\x84' => [ 0xE3, 0x81, 0x82, 0xE3, 0x81, 0x84 ]
// r'\343\201\202\343\201\204' => [ 0343, 0201, 0202, 0343, 0201, 0204 ]
function convertEscapedBytesToBytes(str, base) {
  var prefix = (base == 16 ? "\\x" : "\\");
  return convertEscapedCodesToCodes(str, prefix, base, 8);
}

// "&amp;#12354;&amp;#12356;" => [ 0x3042, 0x3044 ]
// "&amp;#x3042;&amp;#x3044;" => [ 0x3042, 0x3044 ]
function convertNumRefToUnicodeCodePoints(str, base) {
  var num_refs = str.split(";");
  num_refs.pop();  // Trim the last element.
  var unicode_codes = [];
  for (var i = 0; i < num_refs.length; ++i) {
    var decimal_str = num_refs[i].replace(/^&#x?/, "");
    var unicode_code = parseInt(decimal_str, base);
    unicode_codes.push(unicode_code);
  }
  return unicode_codes;
}

// [ 0x3042, 0x3044 ] => [ 0x3042, 0x3044 ]
// [ 0xD840, 0xDC0B ] => [ 0x2000B ]  // A surrogate pair.
function convertUnicodeCodePointsToUtf16Codes(unicode_codes) {
  var utf16_codes = [];
  for (var i = 0; i < unicode_codes.length; ++i) {
    var unicode_code = unicode_codes[i];
    if (unicode_code < (1 << 16)) {
      utf16_codes.push(unicode_code);
    } else {
      var first = ((unicode_code - (1 << 16)) / (1 << 10)) + 0xD800;
      var second = (unicode_code % (1 << 10)) + 0xDC00;
      utf16_codes.push(first)
      utf16_codes.push(second)
    }
  }
  return utf16_codes;
}

// 0x3042 => [ 0xE3, 0x81, 0x82 ]
function convertUnicodeCodePointToUtf8Bytes(unicode_code, base) {
  var utf8_bytes = [];
  if (unicode_code < 0x80) {  // 1-byte
    utf8_bytes.push(unicode_code);
  } else if (unicode_code < (1 << 11)) {  // 2-byte
    utf8_bytes.push((unicode_code >>> 6) | 0xC0);
    utf8_bytes.push((unicode_code & 0x3F) | 0x80);
  } else if (unicode_code < (1 << 16)) {  // 3-byte
    utf8_bytes.push((unicode_code >>> 12) | 0xE0);
    utf8_bytes.push(((unicode_code >> 6) & 0x3f) | 0x80);
    utf8_bytes.push((unicode_code & 0x3F) | 0x80);
  } else if (unicode_code < (1 << 21)) {  // 4-byte
    utf8_bytes.push((unicode_code >>> 18) | 0xF0);
    utf8_bytes.push(((unicode_code >> 12) & 0x3F) | 0x80);
    utf8_bytes.push(((unicode_code >> 6) & 0x3F) | 0x80);
    utf8_bytes.push((unicode_code & 0x3F) | 0x80);
  }
  return utf8_bytes;
}

// [ 0x3042, 0x3044 ] => [ 0xE3, 0x81, 0x82, 0xE3, 0x81, 0x84 ]
function convertUnicodeCodePointsToUtf8Bytes(unicode_codes) {
  var utf8_bytes = [];
  for (var i = 0; i < unicode_codes.length; ++i) {
    var bytes = convertUnicodeCodePointToUtf8Bytes(unicode_codes[i]);
    utf8_bytes = utf8_bytes.concat(bytes);
  }
  return utf8_bytes;
}

// 0xff => "ff"
// 0xff => "377"
function formatNumber(number, base, num_digits) {
  var str = number.toString(base).toUpperCase();
  for (var i = str.length; i < num_digits; ++i) {
    str = "0" + str;
  }
  return str;
}

var BASE64 =
   "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/";

function encodeBase64Helper(data) {
  var encoded = [];
  if (data.length == 1) {
    encoded.push(BASE64.charAt(data[0] >> 2));
    encoded.push(BASE64.charAt(((data[0] & 3) << 4)));
    encoded.push('=');
    encoded.push('=');
  } else if (data.length == 2) {
    encoded.push(BASE64.charAt(data[0] >> 2));
    encoded.push(BASE64.charAt(((data[0] & 3) << 4) |
                               (data[1] >> 4)));
    encoded.push(BASE64.charAt(((data[1] & 0xF) << 2)));
    encoded.push('=');
  } else if (data.length == 3) {
    encoded.push(BASE64.charAt(data[0] >> 2));
    encoded.push(BASE64.charAt(((data[0] & 3) << 4) |
                               (data[1] >> 4)));
    encoded.push(BASE64.charAt(((data[1] & 0xF) << 2) |
                               (data[2] >> 6)));
    encoded.push(BASE64.charAt(data[2] & 0x3f));
  }
  return encoded.join('');
}

// "44GC44GE" => [ 0xE3, 0x81, 0x82, 0xE3, 0x81, 0x84 ]
function decodeBase64(encoded) {
  var decoded_bytes = [];
  var data_bytes = [];
  for (var i = 0; i < encoded.length; i += 4) {
    data_bytes.length = 0;
    for (var j = i; j < i + 4; ++j) {
      var letter = encoded.charAt(j);
      if (letter == "=" || letter == "") {
        break;
      }
      var data_byte = BASE64.indexOf(letter);
      if (data_byte >= 64) {  // Malformed base64 data.
        break;
      }
      data_bytes.push(data_byte);
    }
    if (data_bytes.length == 1) {
      // Malformed base64 data.
    } else if (data_bytes.length == 2) {  // 12-bit.
      decoded_bytes.push((data_bytes[0] << 2) | (data_bytes[1] >> 4));
    } else if (data_bytes.length == 3) {  // 18-bit.
      decoded_bytes.push((data_bytes[0] << 2) | (data_bytes[1] >> 4));
      decoded_bytes.push(((data_bytes[1] & 0xF) << 4) | (data_bytes[2] >> 2));
    } else if (data_bytes.length == 4) {  // 24-bit.
      decoded_bytes.push((data_bytes[0] << 2) | (data_bytes[1] >> 4));
      decoded_bytes.push(((data_bytes[1] & 0xF) << 4) | (data_bytes[2] >> 2));
      decoded_bytes.push(((data_bytes[2] & 0x3) << 6) | (data_bytes[3]));
    }
  }
  return decoded_bytes;
}

// [ 0xE3, 0x81, 0x82, 0xE3, 0x81, 0x84 ] => "44GC44GE"
function encodeBase64(data_bytes) {
  var encoded = '';
  for (var i = 0; i <  data_bytes.length; i += 3) {
    var at_most_three_bytes = data_bytes.slice(i, i + 3);
    encoded += encodeBase64Helper(at_most_three_bytes);
  }
  return encoded;
}

function decodeQuotedPrintableHelper(str, prefix) {
  var decoded_bytes = [];
  for (var i = 0; i < str.length;) {
    if (str.charAt(i) == prefix) {
      decoded_bytes.push(parseInt(str.substr(i + 1, 2), 16));
      i += 3;
    } else {
      decoded_bytes.push(str.charCodeAt(i));
      ++i;
    }
  }
  return decoded_bytes;
}

// "=E3=81=82=E3=81=84" => [ 0xE3, 0x81, 0x82, 0xE3, 0x81, 0x84 ]
function decodeQuotedPrintable(str) {
  str = str.replace(/_/g, " ")  // RFC 2047.
  return decodeQuotedPrintableHelper(str, "=");
}

// "=E3=81=82=E3=81=84" => [ 0xE3, 0x81, 0x82, 0xE3, 0x81, 0x84 ]
function decodeQuotedPrintableWithoutRFC2047(str) {
  return decodeQuotedPrintableHelper(str, "=");
}

// "%E3%81%82%E3%81%84" => [ 0xE3, 0x81, 0x82, 0xE3, 0x81, 0x84 ]
function decodeUrl(str) {
  return decodeQuotedPrintableHelper(str, "%");
}

function encodeQuotedPrintableHelper(data_bytes, prefix, should_escape) {
  var encoded = '';
  var prefix_code = prefix.charCodeAt(0);
  for (var i = 0; i <  data_bytes.length; ++i) {
    var data_byte = data_bytes[i];
    if (should_escape(data_byte)) {
      encoded += prefix + formatNumber(data_bytes[i], 16, 2);
    } else {
      encoded += String.fromCharCode(data_byte);
    }
  }
  return encoded;
}

// [ 0xE3, 0x81, 0x82, 0xE3, 0x81, 0x84 ] => "=E3=81=82=E3=81=84"
function encodeQuotedPrintable(data_bytes) {
  var should_escape = function(b) {
    return b < 32 || b > 126 || b == EQUAL_SIGN || b == QUESTION_MARK;
  };
  return encodeQuotedPrintableHelper(data_bytes, '=', should_escape);
}

var URL_SAFE =
    "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789_.-";

// [ 0xE3, 0x81, 0x82, 0xE3, 0x81, 0x84 ] => "%E3%81%82%E3%81%84"
function encodeUrl(data_bytes) {
  var should_escape = function(b) {
    return URL_SAFE.indexOf(String.fromCharCode(b)) == -1;
  };
  return encodeQuotedPrintableHelper(data_bytes, '%', should_escape);
}

// [ 0x3042, 0x3044 ] => "あい"
function convertUtf16CodesToString(utf16_codes) {
  var unescaped = '';
  for (var i = 0; i < utf16_codes.length; ++i) {
    unescaped += String.fromCharCode(utf16_codes[i]);
  }
  return unescaped;
}

// [ 0x3042, 0x3044 ] => "あい"
function convertUnicodeCodePointsToString(unicode_codes) {
  var utf16_codes = convertUnicodeCodePointsToUtf16Codes(unicode_codes);
  return convertUtf16CodesToString(utf16_codes);
}

function maybeInitMaps(encoded_maps, to_unicode_map, from_unicode_map) {
  if (to_unicode_map.is_initialized) {
    return;
  }
  var data_types = [ 'ROUNDTRIP', 'INPUT_ONLY', 'OUTPUT_ONLY' ];
  for (var i = 0; i < data_types.length; ++i) {
    var data_type = data_types[i];
    var encoded_data = encoded_maps[data_type];
    var data_bytes = decodeBase64(encoded_data);
    for (var j = 0; j < data_bytes.length; j += 4) {
      var local_code = (data_bytes[j] << 8) | data_bytes[j + 1];
      var unicode_code = (data_bytes[j + 2] << 8) | data_bytes[j + 3];
      if (i == 0 || i == 1) {  // ROUNDTRIP or INPUT_ONLY
        to_unicode_map[local_code] = unicode_code;
      }
      if (i == 0 || i == 2) {  // ROUNDTRIP or OUTPUT_ONLY
        from_unicode_map[unicode_code] = local_code;
      }
    }
  }
  to_unicode_map.is_initialized = true;
}

var SJIS_TO_UNICODE = {}
var UNICODE_TO_SJIS = {}
// Requires: sjis_map.js should be loaded.
function maybeInitSjisMaps() {
  maybeInitMaps(SJIS_MAP_ENCODED, SJIS_TO_UNICODE, UNICODE_TO_SJIS);
}

var ISO88591_TO_UNICODE = {}
var UNICODE_TO_ISO88591 = {}
// Requires: iso88591_map.js should be loaded.
function maybeInitIso88591Maps() {
  maybeInitMaps(ISO88591_MAP_ENCODED, ISO88591_TO_UNICODE,
                UNICODE_TO_ISO88591);
}

function lookupMapWithDefault(map, key, default_value) {
  var value = map[key];
  if (!value) {
    value = default_value;
  }
  return value;
}

// [ 0x3042, 0x3044 ] => [ 0x82, 0xA0, 0x82, 0xA2 ]
function convertUnicodeCodePointsToSjisBytes(unicode_codes) {
  maybeInitSjisMaps();
  var sjis_bytes = [];
  for (var i = 0; i < unicode_codes.length; ++i) {
    var unicode_code = unicode_codes[i];
    var sjis_code = lookupMapWithDefault(UNICODE_TO_SJIS,
                                         unicode_code, QUESTION_MARK);
    if (sjis_code <= 0xFF) { // 1-byte character.
      sjis_bytes.push(sjis_code);
    } else {
      sjis_bytes.push(sjis_code >> 8);
      sjis_bytes.push(sjis_code & 0xFF);
    }
  }
  return sjis_bytes;
}

// [ 0x3042, 0x3044 ] => [ 0xA4, 0xA2, 0xA4, 0xA4 ]
function convertUnicodeCodePointsToEucJpBytes(unicode_codes) {
  maybeInitSjisMaps();
  var eucjp_bytes = [];
  for (var i = 0; i < unicode_codes.length; ++i) {
    var unicode_code = unicode_codes[i];
    var sjis_code = lookupMapWithDefault(UNICODE_TO_SJIS, unicode_code,
                                         QUESTION_MARK);
    if (sjis_code > 0xFF) {  // Double byte character.
      var jis_code = convertSjisCodeToJisX208Code(sjis_code);
      var eucjp_code = jis_code | 0x8080;
      eucjp_bytes.push(eucjp_code >> 8);
      eucjp_bytes.push(eucjp_code & 0xFF);
    } else if (sjis_code >= 0x80) {  // 8-bit character.
      eucjp_bytes.push(0x8E);
      eucjp_bytes.push(sjis_code);
    } else {  // 7-bit character.
      eucjp_bytes.push(sjis_code);
    }
  }
  return eucjp_bytes;
}


function convertUnicodeCodePointsToIso88591Bytes(unicode_codes) {
  maybeInitIso88591Maps();
  var latin_bytes = [];
  for (var i = 0; i < unicode_codes.length; ++i) {
    var unicode_code = unicode_codes[i];
    var latin_code = lookupMapWithDefault(UNICODE_TO_ISO88591,
                                          unicode_code, QUESTION_MARK);
    latin_bytes.push(latin_code);
  }
  return latin_bytes;
}

// [ 0x82, 0xA0, 0x82, 0xA2 ] => [ 0x3042, 0x3044 ]
function convertSjisBytesToUnicodeCodePoints(sjis_bytes) {
  maybeInitSjisMaps();
  var unicode_codes = [];
  for (var i = 0; i < sjis_bytes.length;) {
    var sjis_code = -1;
    var sjis_byte = sjis_bytes[i];
    if ((sjis_byte >= 0x81 && sjis_byte <= 0x9F) ||
        (sjis_byte >= 0xE0 && sjis_byte <= 0xFC)) {
      ++i;
      var sjis_byte2 = sjis_bytes[i];
      if ((sjis_byte2 >= 0x40 && sjis_byte2 <= 0x7E) ||
          (sjis_byte2 >= 0x80 && sjis_byte2 <= 0xFC)) {
        sjis_code = (sjis_byte << 8) | sjis_byte2;
        ++i;
      }
    } else {
      sjis_code = sjis_byte;
      ++i;
    }

    var unicode_code = lookupMapWithDefault(SJIS_TO_UNICODE,
                                            sjis_code, QUESTION_MARK);
    unicode_codes.push(unicode_code);
  }
  return unicode_codes;
}

function convertIso88591BytesToUnicodeCodePoints(latin_bytes) {
  maybeInitIso88591Maps();
  var unicode_codes = [];
  for (var i = 0; i < latin_bytes.length; ++i) {
    var latin_code = latin_bytes[i];
    var unicode_code = lookupMapWithDefault(ISO88591_TO_UNICODE,
                                            latin_code, QUESTION_MARK);
    unicode_codes.push(unicode_code);
  }
  return unicode_codes;
}

// 0x2422 => 0x82a0
function convertJisX208CodeToSjisCode(jis_code) {
  var j1 = jis_code >> 8;
  var j2 = jis_code & 0xFF;
  // http://people.debian.org/~kubota/unicode-symbols-map2.html.ja
  var s1 = ((j1 - 1) >> 1) + ((j1 <= 0x5E) ? 0x71 : 0xB1);
  var s2 = j2 + ((j1 & 1) ? ((j2 < 0x60) ? 0x1F : 0x20) : 0x7E);
  return (s1 << 8) | s2;
}

// 0x82a0 => 0x2422
function convertSjisCodeToJisX208Code(sjis_code) {
  var s1 = sjis_code >> 8;
  var s2 = sjis_code & 0xFF;
  // http://people.debian.org/~kubota/unicode-symbols-map2.html.ja
  var j1 = (s1 << 1) - (s1 <= 0x9f ? 0xe0 : 0x160) - (s2 < 0x9f ? 1 : 0);
  var j2 = s2 - 0x1f - (s2 >= 0x7f ? 1 : 0) - (s2 >= 0x9f ? 0x5e : 0);
  return (j1 << 8) | j2;
}

// [ 0x24, 0x22, 0x24, 0x24 ] => [ 0x82, 0xA0, 0x82, 0xA2 ]
function convertJisX208BytesToSjisBytes(jis_bytes) {
  var sjis_bytes = [];
  for (var i = 0; i < jis_bytes.length; i += 2) {
    var jis_code = (jis_bytes[i] << 8) | jis_bytes[i + 1];
    var sjis_code = convertJisX208CodeToSjisCode(jis_code);
    sjis_bytes.push(sjis_code >> 8);
    sjis_bytes.push(sjis_code & 0xFF);
  }
  return sjis_bytes;
}

// [ 0x82, 0xA0, 0x82, 0xA2 ] => [ 0x24, 0x22, 0x24, 0x24 ]
function convertSjisBytesToJisX208Bytes(sjis_bytes) {
  var jis_bytes = [];
  for (var i = 0; i < sjis_bytes.length; i += 2) {
    var sjis_code = (sjis_bytes[i] << 8) | sjis_bytes[i + 1];
    var jis_code = convertSjisCodeToJisX208Code(sjis_code);
    jis_bytes.push(jis_code >> 8);
    jis_bytes.push(jis_code & 0xFF);
  }
  return jis_bytes;
}

// Constants used in convertJisBytesToUnicodeCodePoints().
var ASCII = 0;
var JISX201 = 1;
var JISX208 = 2;

// Map used in convertIso2022JpBytesToUnicodeCodePoints().
var ESCAPE_SEQUENCE_TO_MODE = {
  "(B": ASCII,
  "(J": JISX201,
  "$B": JISX208,
  "$@": JISX208
};

// Map used in convertUnicodeCodePointsToIso2022JpBytes().
var MODE_TO_ESCAPE_SEQUENCE = {}
MODE_TO_ESCAPE_SEQUENCE[ASCII] = "(B";
MODE_TO_ESCAPE_SEQUENCE[JISX201] = "(J";
MODE_TO_ESCAPE_SEQUENCE[JISX208] = "$B";

// [ 0x1B, 0x24, 0x42, 0x24, 0x22, 0x1B, 0x28, 0x42, ] => [ 0x3042 ]
function convertIso2022JpBytesToUnicodeCodePoints(iso2022jp_bytes) {
  maybeInitSjisMaps();
  var flush = function(mode, data_bytes, output) {
    var unicode_codes = [];
    if (mode == ASCII) {
      unicode_codes = data_bytes;
    } else if (mode == JISX201) {  // Might have half-width Katakana?
      unicode_codes = convertSjisBytesToUnicodeCodePoints(data_bytes);
    } else if (mode == JISX208) {
      var sjis_bytes = convertJisX208BytesToSjisBytes(data_bytes);
      unicode_codes = convertSjisBytesToUnicodeCodePoints(sjis_bytes);
    } else {  // Unknown mode
    }
    for (var i = 0; i < unicode_codes.length; ++i) {
      output.push(unicode_codes[i]);
    }
    data_bytes.length = 0;  // Clear.
  }

  var unicode_codes = [];
  var mode = ASCII;
  var current_data_bytes = [];
  for (var i = 0; i < iso2022jp_bytes.length;) {
    if (iso2022jp_bytes[i] == 0x1B) {  // Mode is changed.
      flush(mode, current_data_bytes, unicode_codes);
      ++i;
      var code = String.fromCharCode(iso2022jp_bytes[i],
                                     iso2022jp_bytes[i + 1]);
      mode = ESCAPE_SEQUENCE_TO_MODE[code];
      if (!mode) {  // Unknown mode.
        mode = ASCII;
      }
      i += 2;
    } else {
      current_data_bytes.push(iso2022jp_bytes[i]);
      ++i;
    }
  }
  flush(mode, current_data_bytes, unicode_codes);
  return unicode_codes;
}

// [ 0xA4, 0xA2, 0xA4, 0xA4 ] => [ 0x3042, 0x3044 ]
function convertEucJpBytesToUnicodeCodePoints(eucjp_bytes) {
  maybeInitSjisMaps();
  var unicode_codes = [];
  for (var i = 0; i < eucjp_bytes.length;) {
    if (eucjp_bytes[i] >= 0x80 && (i + 1) < eucjp_bytes.length &&
        eucjp_bytes[i + 1] >= 0x80) {
      var eucjp_code = (eucjp_bytes[i] << 8) | eucjp_bytes[i + 1];
      var jis_code = eucjp_code & 0x7F7F;
      var sjis_code = convertJisX208CodeToSjisCode(jis_code);
      var unicode_code = lookupMapWithDefault(SJIS_TO_UNICODE,
                                              sjis_code, QUESTION_MARK);
      unicode_codes.push(unicode_code);
      i += 2;
    } else {
      if (eucjp_bytes[i] < 0x80) {
        unicode_codes.push(eucjp_bytes[i]);
      } else {
        // Ignore singleton 8-bit byte.
      }
      ++i;
    }
  }
  return unicode_codes;
}

//  [ 0x3042 ] => [ 0x1B, 0x24, 0x42, 0x24, 0x22, 0x1B, 0x28, 0x42, ]
function convertUnicodeCodePointsToIso2022JpBytes(unicode_codes) {
  maybeInitSjisMaps();
  var mode = ASCII;
  var maybeChangeMode = function(new_mode) {
    if (mode != new_mode) {
      mode = new_mode;
      var esc_as_string = MODE_TO_ESCAPE_SEQUENCE[mode];
      var esc_as_code_points = convertStringToUnicodeCodePoints(esc_as_string);
      iso2022jp_bytes.push(0x1B);  // ESC code.
      iso2022jp_bytes = iso2022jp_bytes.concat(esc_as_code_points);
    }
  }
  var iso2022jp_bytes = [];
  for (var i = 0; i < unicode_codes.length; ++i) {
    var unicode_code = unicode_codes[i];
    var sjis_code = lookupMapWithDefault(UNICODE_TO_SJIS, unicode_code,
                                         QUESTION_MARK);
    if (sjis_code > 0xFF) {  // Double byte character.
      var jis_code = convertSjisCodeToJisX208Code(sjis_code);
      maybeChangeMode(JISX208);
      iso2022jp_bytes.push(jis_code >> 8);
      iso2022jp_bytes.push(jis_code & 0xFF);
    } else if (sjis_code >= 0x80) {  // 8-bit character.
      maybeChangeMode(JISX201);
      iso2022jp_bytes.push(sjis_code);
    } else {  // 7-bit character.
      maybeChangeMode(ASCII);
      iso2022jp_bytes.push(sjis_code);
    }
  }
  maybeChangeMode(ASCII);
  return iso2022jp_bytes;
}

var MIME_FULL_MATCH = /^=\?([^?]+)\?([BQbq])\?([^?]+)\?=$/;
var MIME_PARTIAL_MATCH = /^=\?([^?]+)\?([BQbq])\?([^?]+)\?=/;

// "=?UTF-8?B?44GC?=" => true
// "foobar" => false
function isMimeEncodedString(str) {
  return str.match(MIME_FULL_MATCH) != null;
}

// "=?UTF-8?B?44GC?=" => ["UTF-8", [0xE3, 0x81, 0x82]]
// "=?UTF-8?Q?=E3=81=82?=" => ["UTF-8", [0xE3, 0x81, 0x82]]
// "INVALID" => []
function decodeMime(str) {
  var m = str.match(MIME_FULL_MATCH);
  if (m) {
    var char_encoding = m[1];
    // We don't need the language information preceded by '*'.
    char_encoding = char_encoding.replace(/\*.*$/, "")
    var mime_encoding = m[2];
    var mime_str = m[3];
    var decoded_bytes;
    if (mime_encoding == "B" || mime_encoding == "b") {
      decoded_bytes = decodeBase64(mime_str);
    } else if (mime_encoding == "Q" || mime_encoding == "q") {
      decoded_bytes = decodeQuotedPrintable(mime_str);
    }
    if (char_encoding != "" && decoded_bytes) {
      return [char_encoding, decoded_bytes]
    }
  }
  return [];
}

var OUTPUT_CONVERTERS = {
  'ISO2022JP': convertUnicodeCodePointsToIso2022JpBytes,
  'ISO88591':  convertUnicodeCodePointsToIso88591Bytes,
  'SHIFTJIS':  convertUnicodeCodePointsToSjisBytes,
  'EUCJP':     convertUnicodeCodePointsToEucJpBytes,
  'UTF8':      convertUnicodeCodePointsToUtf8Bytes
}

var INPUT_CONVERTERS = {
  'ISO2022JP': convertIso2022JpBytesToUnicodeCodePoints,
  'ISO88591':  convertIso88591BytesToUnicodeCodePoints,
  'SHIFTJIS':  convertSjisBytesToUnicodeCodePoints,
  'EUCJP':     convertEucJpBytesToUnicodeCodePoints,
  'UTF8':      convertUtf8BytesToUnicodeCodePoints
}

function convertUnicodeCodePointsToBytes(unicode_codes, encoding) {
  var normalized_encoding = normalizeEncodingName(encoding);
  var convert_function = OUTPUT_CONVERTERS[normalized_encoding];
  if (convert_function) {
    return convert_function(unicode_codes);
  }
  return [];
}

function convertBytesToUnicodeCodePoints(data_bytes, encoding) {
  var normalized_encoding = normalizeEncodingName(encoding);
  var convert_function = INPUT_CONVERTERS[normalized_encoding];
  if (convert_function) {
    return convert_function(data_bytes);
  }
  return [];
}

// 'あい' => r'\u3042\u3044'
function escapeToUtf16(str) {
  var escaped = ''
  for (var i = 0; i < str.length; ++i) {
    var hex = str.charCodeAt(i).toString(16).toUpperCase();
    escaped += "\\u" + "0000".substr(hex.length) + hex;
  }
  return escaped;
}

// 'あい' => r'\U00003042\U00003044'
function escapeToUtf32(str) {
  var escaped = ''
  var unicode_codes = convertStringToUnicodeCodePoints(str);
  for (var i = 0; i <unicode_codes.length; ++i) {
    var hex = unicode_codes[i].toString(16).toUpperCase();
    escaped += "\\U" + "00000000".substr(hex.length) + hex;
  }
  return escaped;
}

// "あい" => "&#12354;&#12356;"
// "あい" => "&#x3042;&#x3044;"
function escapeToNumRef(str, base) {
  var unicode_codes = convertStringToUnicodeCodePoints(str);
  var escaped = ''
  var prefix = base == 10 ? ''  : 'x';
  for (var i = 0; i < unicode_codes.length; ++i) {
    var code = unicode_codes[i].toString(base).toUpperCase();
    var num_ref = "&#" + prefix + code + ";"
    escaped += num_ref;
  }
  return escaped;
}

// "あい" => "l8je"
function escapeToPunyCode(str) {
  var unicode_codes = convertStringToPunyCodes(str);
  return convertUnicodeCodePointsToString(unicode_codes);
}

// [ 0xE3, 0x81, 0x82, 0xE3, 0x81, 0x84 ] => '\xE3\x81\x82\xE3\x81\x84'
// [ 0343, 0201, 0202, 0343, 0201, 0204 ] => '\343\201\202\343\201\204'
function convertBytesToEscapedString(data_bytes, base) {
  var escaped = '';
  for (var i = 0; i < data_bytes.length; ++i) {
    var prefix = (base == 16 ? "\\x" : "\\");
    var num_digits = base == 16 ? 2 : 3;
    var escaped_byte = prefix + formatNumber(data_bytes[i], base, num_digits)
    escaped += escaped_byte;
  }
  return escaped;
}

// "あい" => [0x6C, 0x38, 0x6A, 0x65]  // "l8je"
// Requires: punycode.js should be loaded.
function convertStringToPunyCodes(str) {
  var unicode_codes = convertStringToUnicodeCodePoints(str);
  var puny_codes = [];
  var result = "";
  if (PunyCode.encode(unicode_codes, puny_codes)) {
    return puny_codes;
  }
  return unicode_codes;
}

// [ 0x6C, 0x38, 0x6A, 0x65 ] => "あい"
// Requires: punycode.js should be loaded.
function convertPunyCodesToString(puny_codes) {
  var unicode_codes = [];
  if (PunyCode.decode(puny_codes, unicode_codes)) {
    return convertUnicodeCodePointsToString(unicode_codes);
  }
  return convertUnicodeCodePointsToString(puny_codes);
}

// "あい" => r'\xE3\x81\x82\xE3\x81\x84'  // UTF-8
// "あい" => r'\343\201\202\343\201\204'  // UTF-8
function escapeToEscapedBytes(str, base, encoding) {
  var unicode_codes = convertStringToUnicodeCodePoints(str);
  var data_bytes = convertUnicodeCodePointsToBytes(unicode_codes, encoding);
  return convertBytesToEscapedString(data_bytes, base);
}

// "あい" => "44GC44GE"  // UTF-8
function escapeToBase64(str, encoding) {
  var unicode_codes = convertStringToUnicodeCodePoints(str);
  var data_bytes = convertUnicodeCodePointsToBytes(unicode_codes, encoding);
  return encodeBase64(data_bytes);
}

// "あい" => "=E3=81=82=E3=81=84"  // UTF-8
function escapeToQuotedPrintable(str, encoding) {
  var unicode_codes = convertStringToUnicodeCodePoints(str);
  var data_bytes = convertUnicodeCodePointsToBytes(unicode_codes, encoding);
  return encodeQuotedPrintable(data_bytes);
}

// "あい" => "%E3%81%82%E3%81%84"
function escapeToUrl(str, encoding) {
  var unicode_codes = convertStringToUnicodeCodePoints(str);
  var data_bytes = convertUnicodeCodePointsToBytes(unicode_codes, encoding);
  return encodeUrl(data_bytes);
}

// "あい" => "=?UTF-8?B?44GC44GE?="
// "あい" => "=?UTF-8?Q?=E3=81=82=E3=81=84?="
function escapeToMime(str, mime_encoding, char_encoding) {
  var unicode_codes = convertStringToUnicodeCodePoints(str);
  var data_bytes = convertUnicodeCodePointsToBytes(unicode_codes,
                                                   char_encoding);
  if (str == "") {
    return "";
  }
  var escaped = "=?" + char_encoding + "?";
  if (mime_encoding == 'base64') {
    escaped += "B?";
    escaped += encodeBase64(data_bytes);
  } else {
    escaped += "Q?";
    escaped += encodeQuotedPrintable(data_bytes);
  }
  escaped += '?=';
  return escaped;
}

// r'\u3042\u3044 => "あい"
function unescapeFromUtf16(str) {
  var utf16_codes = convertEscapedUtf16CodesToUtf16Codes(str);
  return convertUtf16CodesToString(utf16_codes);
}

// r'\U00003042\U00003044 => "あい"
function unescapeFromUtf32(str) {
  var unicode_codes = convertEscapedUtf32CodesToUnicodeCodePoints(str);
  var utf16_codes = convertUnicodeCodePointsToUtf16Codes(unicode_codes);
  return convertUtf16CodesToString(utf16_codes);
}

// r'\xE3\x81\x82\xE3\x81\x84' => "あい"
// r'\343\201\202\343\201\204' => "あい"
function unescapeFromEscapedBytes(str, base, encoding) {
  var data_bytes = convertEscapedBytesToBytes(str, base);
  var unicode_codes = convertBytesToUnicodeCodePoints(data_bytes, encoding);
  return convertUnicodeCodePointsToString(unicode_codes);
}


// "&#12354;&#12356;" => "あい"
// "&#x3042;&#x3044;" => "あい"
function unescapeFromNumRef(str, base) {
  var unicode_codes = convertNumRefToUnicodeCodePoints(str, base);
  return convertUnicodeCodePointsToString(unicode_codes);
}

// "l8je" => "あい"
function unescapeFromPunyCode(str) {
  var unicode_codes = convertStringToUnicodeCodePoints(str);
  return convertPunyCodesToString(unicode_codes);
}

// "44GC44GE" => "あい"
function unescapeFromBase64(str, encoding) {
  var decoded_bytes = decodeBase64(str);
  var unicode_codes = convertBytesToUnicodeCodePoints(decoded_bytes, encoding);
  return convertUnicodeCodePointsToString(unicode_codes);
}

// "=E3=81=82=E3=81=84" => "あい"
function unescapeFromQuotedPrintable(str, encoding) {
  var decoded_bytes = decodeQuotedPrintable(str);
  var unicode_bytes = convertBytesToUnicodeCodePoints(decoded_bytes, encoding);
  return convertUnicodeCodePointsToString(unicode_bytes);
}

function unescapeFromQuotedPrintableWithoutRFC2047(str, encoding) {
  var decoded_bytes = decodeQuotedPrintableWithoutRFC2047(str);
  var unicode_bytes = convertBytesToUnicodeCodePoints(decoded_bytes, encoding);
  return convertUnicodeCodePointsToString(unicode_bytes);
}

// "%E3%81%82%E3%81%84" => "あい"
function unescapeFromUrl(str, encoding) {
  var decoded_bytes = decodeUrl(str);
  var unicode_bytes = convertBytesToUnicodeCodePoints(decoded_bytes, encoding);
  return convertUnicodeCodePointsToString(unicode_bytes);
}

// " " => true
// " \n" => true
function isEmptyOrSequenceOfWhiteSpaces(str) {
  for (var i = 0; i < str.length; ++i) {
    var code = str.charCodeAt(i);
    if (!(code == 0x09 ||   // TAB
          code == 0x0A ||   // LF
          code == 0x0D ||   // CR
          code == 0x20)) {  // SPACE
      return false;
    }
  }
  return true;
}

// "=?UTF-8?B?*?= =?UTF-8?B?*?=" => ["=?UTF-8?B?*?=", "=?UTF-8?B?*?="]
// "=?UTF-8?B?*?=FOO" => ["=?UTF-8?B?*?=", "FOO"]
function splitMimeString(str) {
  if(!str) { return []; }

  var parts = [];
  var current = "";
  while (str != "") {
    var m = str.match(MIME_PARTIAL_MATCH)
    if (m) {
      if (!isEmptyOrSequenceOfWhiteSpaces(current)) {
        parts.push(current);
      }
      current = "";
      parts.push(m[0]);
      str = str.substr(m[0].length);
    } else {
      current += str.charAt(0);
      str = str.substr(1);
    }
  }
  if (!isEmptyOrSequenceOfWhiteSpaces(current)) {
    parts.push(current);
  }
  return parts;
}

// "UTF-8" => "UTF8"
// "Shift_JIS" => "SHIFTJIS"
function normalizeEncodingName(encoding) {
  return encoding.toUpperCase().replace(/[_-]/g, "");
}

// "=?UTF-8?B?44GC44GE?=" => "あい"
// "=?Shift_JIS?B?gqCCog==?=" => "あい"
// "=?ISO-2022-JP?B?GyRCJCIkJBsoQg==?=" => "あい"
// "=?UTF-8?Q?=E3=81=82=E3=81=84?=" => "あい"
// "=?Shift_JIS?Q?=82=A0=82=A2?=" => "あい"
// "=?ISO-2022-JP?Q?=1B$B$"$$=1B(B?=" => "あい"
function unescapeFromMime(str) {
  var parts = splitMimeString(str);
  var unescaped = "";
  for (var i = 0; i < parts.length; ++i) {
    if (isMimeEncodedString(parts[i])) {
      var pair = decodeMime(parts[i]);
      if (pair.length == 0) {  // Malformed MIME string.  Skip it.
        continue;
      }
      var encoding = normalizeEncodingName(pair[0]);
      var data_bytes = pair[1];
      var unicode_codes = convertBytesToUnicodeCodePoints(data_bytes,
                                                          encoding);
      unescaped += convertUnicodeCodePointsToString(unicode_codes);
    } else {
      unescaped += parts[i];
    }
  }
  return unescaped;
}
