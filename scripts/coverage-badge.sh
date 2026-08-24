#!/usr/bin/env sh
# Generates a shields-style SVG coverage badge on stdout.
# Usage: coverage-badge.sh <percent>   e.g. coverage-badge.sh 84.4
set -eu

PCT="$1"
VALUE="${PCT}%"

# Color thresholds, matching the usual shields.io palette.
INT=$(printf '%.0f' "$PCT")
if [ "$INT" -ge 90 ]; then COLOR="#4c1" # brightgreen
elif [ "$INT" -ge 80 ]; then COLOR="#97ca00" # green
elif [ "$INT" -ge 70 ]; then COLOR="#a4a61d" # yellowgreen
elif [ "$INT" -ge 60 ]; then COLOR="#dfb317" # yellow
elif [ "$INT" -ge 50 ]; then COLOR="#fe7d37" # orange
else COLOR="#e05d44" # red
fi

LABEL="coverage"
LABEL_W=61
VALUE_W=$((${#VALUE} * 7 + 10))
TOTAL_W=$((LABEL_W + VALUE_W))
LABEL_X=$((LABEL_W * 10 / 2))
VALUE_X=$(((LABEL_W + VALUE_W / 2) * 10))
LABEL_TL=$(((LABEL_W - 10) * 10))
VALUE_TL=$(((VALUE_W - 10) * 10))

cat <<EOF
<svg xmlns="http://www.w3.org/2000/svg" width="$TOTAL_W" height="20" role="img" aria-label="$LABEL: $VALUE">
  <title>$LABEL: $VALUE</title>
  <linearGradient id="s" x2="0" y2="100%">
    <stop offset="0" stop-color="#bbb" stop-opacity=".1"/>
    <stop offset="1" stop-opacity=".1"/>
  </linearGradient>
  <clipPath id="r"><rect width="$TOTAL_W" height="20" rx="3" fill="#fff"/></clipPath>
  <g clip-path="url(#r)">
    <rect width="$LABEL_W" height="20" fill="#555"/>
    <rect x="$LABEL_W" width="$VALUE_W" height="20" fill="$COLOR"/>
    <rect width="$TOTAL_W" height="20" fill="url(#s)"/>
  </g>
  <g fill="#fff" text-anchor="middle" font-family="Verdana,Geneva,DejaVu Sans,sans-serif" font-size="110" text-rendering="geometricPrecision">
    <text x="$LABEL_X" y="140" transform="scale(.1)" textLength="$LABEL_TL" fill="#010101" fill-opacity=".3">$LABEL</text>
    <text x="$LABEL_X" y="130" transform="scale(.1)" textLength="$LABEL_TL">$LABEL</text>
    <text x="$VALUE_X" y="140" transform="scale(.1)" textLength="$VALUE_TL" fill="#010101" fill-opacity=".3">$VALUE</text>
    <text x="$VALUE_X" y="130" transform="scale(.1)" textLength="$VALUE_TL">$VALUE</text>
  </g>
</svg>
EOF
