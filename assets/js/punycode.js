// https://raw.githubusercontent.com/jupiter/node-strutil/master/lib/punycode.js

// Almost literal translation of the C code in
// http://www.ietf.org/rfc/rfc3492.txt to JavaScript.
// Comments removed.  Removed "case_flags" support.

function InitPunyCode() {
var BASE = 36;
var TMIN = 1;
var TMAX = 26;
var SKEW = 38;
var DAMP = 700;
var INITIAL_BIAS = 72;
var INITIAL_N = 0x80;
var DELIMITER = 0x2D;
var MAXINT = Math.pow(2, 31) -1;  // In 32-bit signed integer.

function basic(cp) {
  return cp < 0x80;
}

function delim(cp) {
  return cp == DELIMITER;
}

function decode_digit(cp) {
  return  (cp - 48 < 10 ? cp - 22 :  cp - 65 < 26 ? cp - 65 :
           cp - 97 < 26 ? cp - 97 :  BASE);
}

function encode_digit(d, flag) {
  return (d + 22 + 75 * (d < 26) - ((flag != 0) << 5));
}

function flagged(bcp) {
  return (bcp - 65 < 26);
}

function encode_basic(bcp, flag) {
  bcp -= (bcp - 97 < 26) << 5;
  return bcp + ((!flag && (bcp - 65 < 26)) << 5);
}

function adapt(delta, numpoints, firsttime) {
  var k;
  delta = Math.floor(firsttime ? delta / DAMP : delta / 2);
  delta += Math.floor(delta / numpoints);
  for (k = 0;  delta > Math.floor(((BASE - TMIN) * TMAX) / 2);  k += BASE) {
    delta = Math.floor(delta / (BASE - TMIN));
  }
  return Math.floor(k + (BASE - TMIN + 1) * delta / (delta + SKEW));
}

// Encodes "input" in punycode.  On success, returns true
// and stores the result in "output".  Both "input" and
// "output" are arrays of Unicode code points.
function punycode_encode(input, output) {
  var n = INITIAL_N;
  var delta = 0;
  var out = 0;
  var bias = INITIAL_BIAS;

  for (var j = 0;  j < input.length;  ++j) {
    if (basic(input[j])) {
      output.push(input[j]);
    }
  }

  var h = out;
  var b = out;
  if (b > 0) output.push(DELIMITER);

  while (h < input.length) {
    var m = MAXINT;
    for (var j = 0;  j < input.length;  ++j) {
      if (input[j] >= n && input[j] < m) m = input[j];
    }

    if (m - n > Math.floor((MAXINT - delta) / (h + 1)))
      return false;
    delta += (m - n) * (h + 1);
    n = m;

    for (var j = 0;  j < input.length;  ++j) {
      if (input[j] < n /* || basic(input[j]) */ ) {
        if (++delta == 0) return false;
      }

      if (input[j] == n) {
        var q = delta;
        for (var k = BASE;  ;  k += BASE) {
          var t = k <= bias ? TMIN :
              k >= bias + TMAX ? TMAX : k - bias;
          if (q < t) break;
          output.push(encode_digit(t + (q - t) % (BASE - t), 0));
          q = Math.floor((q - t) / (BASE - t));
        }

        output.push(encode_digit(q, false));
        bias = adapt(delta, h + 1, h == b);
        delta = 0;
        ++h;
      }
    }
    ++delta;
    ++n;
  }
  return true;
}

// Decodes "input" to Unicode code points.  On success,
// returns true and stores the result in "output".  Both
// "input" and "output" are arrays of Unicode code points.
function punycode_decode(input, output) {
  var n = INITIAL_N;
  var out = 0;
  var bias = INITIAL_BIAS;

  var b = 0;
  for (var j = 0;  j < input.length;  ++j) if (delim(input[j])) b = j;

  for (var j = 0;  j < b;  ++j) {
    if (!basic(input[j])) return false;
    output.push(input[j]);
  }

  var i = 0;
  var inp = b > 0 ? b + 1 : 0;
  for (; inp < input.length;  ++out) {
    var oldi = i;
    var w = 1;
    for (var k = BASE;  ;  k += BASE) {
      if (inp >= input.length) return false;
      var digit = decode_digit(input[inp++]);
      if (digit >= BASE) return false;
      if (digit > Math.floor((MAXINT - i) / w)) return false;
      i += digit * w;
      var t = (k <= bias ? TMIN :
               k >= bias + TMAX ? TMAX : k - bias);
      if (digit < t) break;
      if (w > Math.floor(MAXINT / (BASE - t))) return false;
      w *= (BASE - t);
    }

    bias = adapt(i - oldi, out + 1, oldi == 0);

    if (i / (out + 1) > MAXINT - n) return false;
    n += Math.floor(i / (out + 1));
    i %= (out + 1);
    output.splice(i, 0, n);
    ++i;
  }
  return true;
}

// Exported functions.
return { encode: punycode_encode,
         decode: punycode_decode };
};

var PunyCode = exports.PunyCode = InitPunyCode();
