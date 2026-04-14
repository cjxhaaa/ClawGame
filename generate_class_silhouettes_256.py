import os

OUT_DIR = "apps/web/public/assets"

width = 256
height = 256

base_head = """
<!-- Head & Hair -->
<rect x="40" y="24" width="48" height="36" fill="#fadcbe" />
"""

commoner_svg = f"""<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 {width} {height}" width="{width}" height="{height}" shape-rendering="crispEdges">
<g transform="scale(2)">
  <!-- Hair: Blonde -->
  <rect x="32" y="16" width="64" height="24" fill="#d9cd7a" />
  <rect x="24" y="32" width="16" height="32" fill="#d9cd7a" />
  <rect x="88" y="32" width="16" height="32" fill="#d9cd7a" />
  
  {base_head}
  <!-- Eyes closed -->
  <rect x="44" y="44" width="12" height="4" fill="#3e2d26" />
  <rect x="72" y="44" width="12" height="4" fill="#3e2d26" />

  <!-- Body: grey purple striped shirt -->
  <rect x="36" y="60" width="56" height="28" fill="#585066" />
  <rect x="36" y="68" width="56" height="4" fill="#443c51" />
  <rect x="36" y="76" width="56" height="4" fill="#443c51" />

  <!-- Skirt: light blue -->
  <rect x="40" y="88" width="48" height="16" fill="#9db0c6" />

  <!-- Legs & Boots -->
  <rect x="44" y="104" width="12" height="8" fill="#fadcbe" />
  <rect x="72" y="104" width="12" height="8" fill="#fadcbe" />
  <rect x="40" y="112" width="20" height="12" fill="#3b2b28" />
  <rect x="68" y="112" width="20" height="12" fill="#3b2b28" />
</g>
</svg>"""

mage_svg = f"""<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 {width} {height}" width="{width}" height="{height}" shape-rendering="crispEdges">
<g transform="scale(2)">
  {base_head}
  <!-- Eyes closed -->
  <rect x="44" y="44" width="12" height="4" fill="#3e2d26" />
  <rect x="72" y="44" width="12" height="4" fill="#3e2d26" />

  <!-- Wizard Hat -->
  <rect x="24" y="20" width="80" height="8" fill="#2d2242" />
  <polygon points="40,20 88,20 64,0" fill="#2d2242" />

  <!-- Dark Robe -->
  <rect x="32" y="60" width="64" height="48" fill="#3a2e5d" />
  <rect x="56" y="60" width="16" height="48" fill="#261b40" />

  <!-- Boots -->
  <rect x="40" y="108" width="20" height="16" fill="#181124" />
  <rect x="68" y="108" width="20" height="16" fill="#181124" />
</g>
</svg>"""

priest_svg = f"""<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 {width} {height}" width="{width}" height="{height}" shape-rendering="crispEdges">
<g transform="scale(2)">
  <!-- Hair: soft gold -->
  <rect x="32" y="16" width="64" height="24" fill="#e8d890" />
  <rect x="24" y="32" width="16" height="32" fill="#e8d890" />
  <rect x="88" y="32" width="16" height="32" fill="#e8d890" />

  {base_head}
  <!-- Eyes closed -->
  <rect x="44" y="44" width="12" height="4" fill="#3e2d26" />
  <rect x="72" y="44" width="12" height="4" fill="#3e2d26" />

  <!-- White Robe -->
  <rect x="32" y="60" width="64" height="48" fill="#f0edf4" />
  <rect x="56" y="60" width="16" height="48" fill="#d9d5e3" />

  <!-- Cross -->
  <rect x="60" y="70" width="8" height="24" fill="#ffd700" />
  <rect x="52" y="76" width="24" height="8" fill="#ffd700" />

  <!-- Boots -->
  <rect x="40" y="108" width="20" height="16" fill="#c4bfd0" />
  <rect x="68" y="108" width="20" height="16" fill="#c4bfd0" />
</g>
</svg>"""

warrior_svg = f"""<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 {width} {height}" width="{width}" height="{height}" shape-rendering="crispEdges">
<g transform="scale(2)">
  {base_head}
  <!-- Eyes closed -->
  <rect x="44" y="44" width="12" height="4" fill="#3e2d26" />
  <rect x="72" y="44" width="12" height="4" fill="#3e2d26" />

  <!-- Helm / Hair -->
  <rect x="32" y="8" width="64" height="24" fill="#7d8896" />
  <rect x="28" y="12" width="8" height="24" fill="#606a77" />
  <rect x="92" y="12" width="8" height="24" fill="#606a77" />
  <rect x="48" y="4" width="32" height="8" fill="#4d545e" />

  <!-- Armor -->
  <rect x="36" y="60" width="56" height="20" fill="#8c97a5" />
  <rect x="32" y="60" width="12" height="20" fill="#687383" />
  <rect x="84" y="60" width="12" height="20" fill="#687383" />
  <rect x="40" y="80" width="48" height="12" fill="#545e6d" />
  <rect x="56" y="68" width="16" height="12" fill="#b0bccd" />

  <!-- Legs Armor -->
  <rect x="44" y="92" width="12" height="16" fill="#3c4652" />
  <rect x="72" y="92" width="12" height="16" fill="#3c4652" />

  <!-- Boots -->
  <rect x="40" y="108" width="20" height="16" fill="#2d3540" />
  <rect x="68" y="108" width="20" height="16" fill="#2d3540" />
</g>
</svg>"""

with open(f"{OUT_DIR}/silhouette-commoner.svg", "w") as f:
    f.write(commoner_svg)
with open(f"{OUT_DIR}/silhouette-mage.svg", "w") as f:
    f.write(mage_svg)
with open(f"{OUT_DIR}/silhouette-priest.svg", "w") as f:
    f.write(priest_svg)
with open(f"{OUT_DIR}/silhouette-warrior.svg", "w") as f:
    f.write(warrior_svg)

print("Generated 4 new 256x256 pixel character files in assets!")
