#!/usr/bin/env python3
"""
Run this script after saving your logo PNG to:
  /Users/chris.robbins/code/mail-trapper/docs/Mailtrapper.png

It will embed the PNG into assets.go as the logo image.
"""
import re, gzip, io, sys

ASSETS_PATH = "/Users/chris.robbins/code/mail-trapper/vendor/github.com/mailhog/MailHog-UI/assets/assets.go"
LOGO_PATH = "/Users/chris.robbins/code/mail-trapper/docs/Mailtrapper.png"

with open(LOGO_PATH, 'rb') as f:
    logo_data = f.read()

if logo_data[:4] != bytes([0x89, 0x50, 0x4e, 0x47]):
    print(f"ERROR: {LOGO_PATH} is not a valid PNG file.")
    sys.exit(1)

print(f"Logo: {len(logo_data)} bytes, valid PNG")

with open(ASSETS_PATH, "rb") as f:
    src = f.read()

def recompress(data):
    out = io.BytesIO()
    with gzip.GzipFile(fileobj=out, mode='wb', mtime=0) as gz:
        gz.write(data)
    return out.getvalue()

def encode_go_bytes(data):
    result = []
    for b in data:
        if b == ord('"'):
            result.append('\\"')
        elif b == ord('\\'):
            result.append('\\\\')
        elif 0x20 <= b < 0x7f:
            result.append(chr(b))
        else:
            result.append(f'\\x{b:02x}')
    return ''.join(result)

def find_var_range(src, var_name_substr):
    var_pattern = rb'var (_[a-zA-Z0-9_]+) = \[\]byte\('
    for m in re.finditer(var_pattern, src):
        var_name = m.group(1).decode()
        if var_name_substr.lower() not in var_name.lower():
            continue
        start = m.end()
        if src[start:start+1] != b'"':
            continue
        end = start + 1
        while end < len(src):
            c = src[end:end+1]
            if c == b'"':
                backslashes = 0
                pos = end - 1
                while pos >= start+1 and src[pos:pos+1] == b'\\':
                    backslashes += 1
                    pos -= 1
                if backslashes % 2 == 0:
                    break
            end += 1
        raw_str = src[start+1:end]
        data = raw_str.decode('unicode_escape').encode('latin-1')
        return gzip.decompress(data), var_name, start, end
    return None, None, None, None

# Find and replace the hog.png asset
content, var_name, start, end = find_var_range(src, 'ImagesHogPng')
if content is None:
    # Already renamed to logo.png in a previous run
    content, var_name, start, end = find_var_range(src, 'ImagesLogoPng')

if content is None:
    print("ERROR: Could not find hog.png or logo.png asset variable")
    sys.exit(1)

print(f"Found image asset: {var_name}")

new_data = recompress(logo_data)
new_encoded = encode_go_bytes(new_data)

# Update size
for old_name in [b'"assets/images/hog.png"', b'"assets/images/logo.png"']:
    size_pat = b'name: ' + old_name + b', size: '
    idx2 = src.find(size_pat)
    if idx2 >= 0:
        num_start = idx2 + len(size_pat)
        num_end = num_start
        while num_end < len(src) and src[num_end:num_end+1].isdigit():
            num_end += 1
        src = src[:num_start] + str(len(logo_data)).encode() + src[num_end:]
        break

# Rename hog.png -> logo.png throughout
src = src.replace(b'"assets/images/hog.png"', b'"assets/images/logo.png"')
src = src.replace(b'name: "assets/images/hog.png"', b'name: "assets/images/logo.png"')
src = src.replace(b'"images/hog.png"', b'"images/logo.png"')

# Re-find after replacements and patch the byte data
content2, var_name2, start2, end2 = find_var_range(src, 'ImagesHogPng')
if content2 is None:
    content2, var_name2, start2, end2 = find_var_range(src, 'ImagesLogoPng')

if content2 is not None:
    src = src[:start2+1] + new_encoded.encode('utf-8') + src[end2:]
    print(f"Logo embedded successfully ({len(logo_data)} bytes -> {len(new_data)} compressed)")

with open(ASSETS_PATH, "wb") as f:
    f.write(src)

print("Done! Logo updated in assets.go")
