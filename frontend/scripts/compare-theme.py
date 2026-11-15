#!/usr/bin/env python3
"""Compare captures from theme-regression.py; write reports beside new captures.

Usage: python3 compare-theme.py /tmp/theme-before /tmp/theme-after
Requires Pillow. Exit 1 means differences need review, not necessarily a defect.
"""

import json
from pathlib import Path
import sys

from PIL import Image, ImageChops


if len(sys.argv) != 3:
    raise SystemExit(__doc__)
before, after = (Path(value) for value in sys.argv[1:])
images = sorted(before.glob('*.png'))
assert images, 'No baseline screenshots found'
for previous_difference in after.glob('difference-*.png'):
    previous_difference.unlink()
pixels, styles = [], []
for baseline in images:
    if baseline.name.startswith('difference-'):
        continue
    current = after / baseline.name
    if not current.exists():
        pixels.append({'page': baseline.stem, 'error': 'missing screenshot'})
        continue
    a = Image.open(baseline).convert('RGB')
    b = Image.open(current).convert('RGB')
    if a.size != b.size:
        pixels.append({'page': baseline.stem, 'error': 'size changed'})
        continue
    difference = ImageChops.difference(a, b)
    # Ignore tiny subpixel rendering noise; keep every material pixel change.
    changed = sum(max(pixel) > 8 for pixel in difference.getdata())
    if changed:
        pixels.append({'page': baseline.stem, 'pixels': changed,
                       'percent': round(100 * changed / (a.width * a.height), 6)})
        difference.save(after / ('difference-' + baseline.name))
    old_nodes = json.loads(baseline.with_suffix('.json').read_text())
    new_nodes = json.loads(current.with_suffix('.json').read_text())
    old_by_path = {node['path']: node for node in old_nodes}
    new_by_path = {node['path']: node for node in new_nodes}
    for path in old_by_path.keys() | new_by_path.keys():
        old, new = old_by_path.get(path), new_by_path.get(path)
        if not old or not new or (old['tag'], old['cls']) != (new['tag'], new['cls']):
            styles.append({'page': baseline.stem, 'path': path, 'error': 'element changed'})
            continue
        for coordinate in ('x', 'y', 'w', 'h'):
            previous, actual = old['box'][coordinate], new['box'][coordinate]
            if abs(previous - actual) > 1 / 64:
                styles.append({'page': baseline.stem, 'path': path, 'class': old['cls'],
                               'property': 'box.' + coordinate, 'before': previous, 'after': actual})
        for pseudo in ('element', '::before', '::after'):
            previous, actual = old.get(pseudo, {}), new.get(pseudo, {})
            for prop in previous.keys() | actual.keys():
                if previous.get(prop) != actual.get(prop):
                    styles.append({'page': baseline.stem, 'path': path,
                                   'class': old['cls'], 'pseudo': pseudo, 'property': prop,
                                   'before': previous.get(prop), 'after': actual.get(prop)})
(after / 'pixel-differences.json').write_text(json.dumps(pixels, indent=2))
(after / 'style-differences.json').write_text(json.dumps(styles, ensure_ascii=False, indent=2))
print(json.dumps({'scenes': len(images), 'images_with_changes': len(pixels),
                  'computed_style_changes': len(styles), 'pixel_changes': pixels}, indent=2))
raise SystemExit(1 if pixels or styles else 0)
