#!/usr/bin/env python3
"""
Generate 64x64 dungeon-themed pixel-art equipment icon SVGs.

Each dungeon has its own color palette applied to the main item shape,
replacing the old 32x32 approach where only the background changed.
Run from repo root:  python tools/generate_equipment_icons.py
"""

import os
import textwrap

OUT_BASE = "apps/web/public/assets/equipment-icons"

# ---------------------------------------------------------------------------
# Dungeon color palettes
# ---------------------------------------------------------------------------
# Each palette key:
#   bg        – outer canvas fill
#   bg2       – subtle inner panel tint
#   frame     – outer border stroke
#   frame2    – inner border stroke
#   d         – darkest material (shadows, depth)
#   m         – mid material (main body)
#   l         – light material (highlights, face)
#   acc       – accent / gem / glow
#   det       – dark detail / cut / void

THEMES = {
    "ancient_catacomb": dict(
        bg="#1c1828", bg2="#2a2040",
        frame="#3f364d", frame2="#7f7ca8",
        d="#3f364d", m="#7f7ca8", l="#c9c7dd",
        acc="#e8d9f5", det="#2a2433",
    ),
    "thorned_hollow": dict(
        bg="#131d0f", bg2="#1e3019",
        frame="#2e5130", frame2="#95c36d",
        d="#2e5130", m="#587f44", l="#95c36d",
        acc="#d2d88e", det="#0e1a0b",
    ),
    "sunscar_warvault": dict(
        bg="#220e06", bg2="#391808",
        frame="#a33a24", frame2="#f5d565",
        d="#8b3218", m="#d75333", l="#f0b54d",
        acc="#f5d565", det="#200c04",
    ),
    "obsidian_spire": dict(
        bg="#0c0b14", bg2="#17141f",
        frame="#272235", frame2="#b8b0df",
        d="#272235", m="#433a61", l="#6d6498",
        acc="#d7d2f1", det="#0c0b14",
    ),
    "starter": dict(
        bg="#e8e2cc", bg2="#f0e9d6",
        frame="#7c735d", frame2="#b7ac8c",
        d="#635b48", m="#8d8268", l="#efe5c8",
        acc="#f7f0d8", det="#3b3427",
    ),
    "unknown": dict(
        bg="#252422", bg2="#302e2b",
        frame="#443f3a", frame2="#cdc6b4",
        d="#443f3a", m="#6d6758", l="#d9d0b7",
        acc="#ede8d7", det="#1a1917",
    ),
}

# ---------------------------------------------------------------------------
# Item shape definitions
# Each entry is a list of (color_key, x, y, w, h) tuples.
# color_key is one of: bg, bg2, frame, frame2, d, m, l, acc, det
# or a literal hex string starting with '#'
# ---------------------------------------------------------------------------

ITEMS: dict[str, list] = {}

# HEAD – knight/warrior helmet
ITEMS["head"] = [
    # crown spike
    ("l",   30,  7,  4,  3),
    ("m",   28, 10,  8,  3),
    # crown band
    ("d",   15, 13, 34,  4),
    ("m",   17, 14, 30,  2),
    # upper visor plate
    ("d",   13, 17, 38,  5),
    # face opening background
    ("l",   15, 22, 34,  9),
    # eye slits
    ("det", 17, 23, 11,  5),
    ("det", 36, 23, 11,  5),
    # nose bridge
    ("m",   29, 23,  6,  7),
    # inner eye highlight
    ("acc", 18, 24,  3,  2),
    ("acc", 37, 24,  3,  2),
    # cheek guards
    ("d",    9, 17,  5, 16),
    ("d",   50, 17,  5, 16),
    ("m",   10, 18,  3, 12),
    ("m",   51, 18,  3, 12),
    # chin bevor
    ("d",   13, 31, 38,  7),
    ("m",   15, 32, 34,  5),
    # vent slits on chin
    ("det", 21, 34,  4,  2),
    ("det", 30, 34,  4,  2),
    ("det", 39, 34,  4,  2),
    # neck guard
    ("d",   17, 38, 30,  5),
    ("m",   19, 39, 26,  3),
    # shoulder base
    ("d",   13, 43, 38,  4),
    # top trim accent
    ("acc", 15, 17, 34,  2),
    # visor divider accent
    ("acc", 13, 31, 38,  2),
]

# CHEST – plate breastplate
ITEMS["chest"] = [
    # shoulder left
    ("d",    8,  8, 14,  8),
    ("m",    9,  9, 12,  6),
    ("l",   10, 10,  8,  3),
    # shoulder right
    ("d",   42,  8, 14,  8),
    ("m",   43,  9, 12,  6),
    ("l",   44, 10,  8,  3),
    # pauldron rivets
    ("det", 12, 13,  2,  2),
    ("det", 46, 13,  2,  2),
    # main breastplate
    ("d",   10, 16, 44, 28),
    ("m",   12, 17, 40, 26),
    ("l",   14, 18, 20, 10),
    # center ridge
    ("d",   30, 17,  4, 27),
    ("m",   31, 17,  2, 27),
    # collar
    ("d",   16, 12, 32,  5),
    ("m",   18, 13, 28,  3),
    # chest rune/sigil
    ("acc", 16, 22, 12,  2),
    ("acc", 16, 26, 12,  2),
    ("acc", 14, 22,  2,  6),
    ("acc", 28, 22,  2,  6),
    # belt
    ("d",   10, 44, 44,  6),
    ("m",   12, 45, 40,  4),
    ("acc", 28, 45,  8,  4),
    # lower tassets
    ("d",   12, 50, 14,  5),
    ("d",   38, 50, 14,  5),
    ("m",   13, 51, 12,  3),
    ("m",   39, 51, 12,  3),
]

# NECKLACE – pendant on a chain
ITEMS["necklace"] = [
    # chain top bar
    ("m",   16,  9, 32,  3),
    ("l",   18, 10, 28,  1),
    # chain left side
    ("m",   13, 12,  5, 12),
    ("d",   14, 12,  3, 12),
    # chain right side
    ("m",   46, 12,  5, 12),
    ("d",   47, 12,  3, 12),
    # chain links left (pixel dots)
    ("acc", 14, 14,  2,  2),
    ("acc", 14, 18,  2,  2),
    ("acc", 14, 22,  2,  2),
    # chain links right
    ("acc", 48, 14,  2,  2),
    ("acc", 48, 18,  2,  2),
    ("acc", 48, 22,  2,  2),
    # pendant mount
    ("d",   22, 24, 20,  4),
    ("m",   24, 24, 16,  4),
    # pendant outer
    ("d",   18, 28, 28, 22),
    ("m",   20, 29, 24, 20),
    # pendant inner gem
    ("l",   24, 32, 16, 12),
    ("acc", 28, 34,  8,  8),
    # gem facets
    ("l",   28, 34,  4,  4),
    ("m",   32, 38,  4,  4),
    # pendant bottom tip
    ("m",   28, 50,  8,  4),
    ("acc", 30, 52,  4,  2),
    # collar clasp
    ("acc", 30,  9,  4,  3),
]

# RING – band ring with center gem
ITEMS["ring"] = [
    # band outer ring (top, bottom, left, right)
    ("d",   16, 16, 32,  6),
    ("d",   16, 42, 32,  6),
    ("d",   10, 22,  8, 20),
    ("d",   46, 22,  8, 20),
    # band inner (hollowed)
    ("bg2", 20, 20, 24,  4),
    ("bg2", 20, 40, 24,  4),
    ("bg2", 14, 24,  6, 16),
    ("bg2", 44, 24,  6, 16),
    # band inner fill (open center)
    ("bg2", 20, 24, 24, 16),
    # band material highlights
    ("m",   17, 17, 30,  4),
    ("m",   17, 43, 30,  4),
    ("m",   11, 23,  6, 18),
    ("m",   47, 23,  6, 18),
    # gem mount
    ("d",   22, 20, 20,  6),
    ("d",   22, 38, 20,  6),
    # gem stone
    ("l",   24, 22, 16, 20),
    ("acc", 26, 24, 12, 16),
    # gem facet lines
    ("m",   32, 22,  2, 20),
    ("l",   26, 24, 12,  4),
    # gem shine
    ("acc", 26, 24,  4,  4),
    # band trim
    ("acc", 17, 22,  4,  2),
    ("acc", 43, 22,  4,  2),
    ("acc", 17, 40,  4,  2),
    ("acc", 43, 40,  4,  2),
]

# BOOTS – paired foot armor
ITEMS["boots"] = [
    # Left boot shaft
    ("d",    8,  8, 18, 24),
    ("m",   10,  9, 14, 22),
    ("l",   10,  9,  8,  8),
    # Left boot foot
    ("d",    6, 32, 22,  8),
    ("m",    8, 33, 20,  6),
    ("l",    8, 33, 10,  3),
    # Left boot toe
    ("d",    6, 40, 24,  5),
    ("m",    8, 41, 22,  3),
    # Left boot heel
    ("d",    6, 45, 10,  4),
    # Left Boot sole
    ("d",    6, 49, 24,  3),
    ("acc",  8, 50, 20,  1),

    # Right boot shaft
    ("d",   38,  8, 18, 24),
    ("m",   40,  9, 14, 22),
    ("l",   44,  9,  8,  8),
    # Right boot foot
    ("d",   36, 32, 22,  8),
    ("m",   36, 33, 20,  6),
    ("l",   46, 33, 10,  3),
    # Right boot toe
    ("d",   34, 40, 24,  5),
    ("m",   34, 41, 22,  3),
    # Right boot heel
    ("d",   48, 45, 10,  4),
    # Right sole
    ("d",   34, 49, 24,  3),
    ("acc", 36, 50, 20,  1),

    # Boot strap / buckle left
    ("acc",  9, 20, 16,  3),
    ("d",   14, 19,  6,  5),
    # Boot strap / buckle right
    ("acc", 39, 20, 16,  3),
    ("d",   44, 19,  6,  5),

    # Knee guard left
    ("d",    8,  8, 18,  4),
    ("acc",  9,  9, 16,  2),
    # Knee guard right
    ("d",   38,  8, 18,  4),
    ("acc", 39,  9, 16,  2),
]

# SWORD + SHIELD
ITEMS["sword_shield"] = [
    # ---- SHIELD (left side) ----
    # outer shield body
    ("d",    6,  6, 22, 30),
    ("m",    8,  7, 18, 28),
    ("l",    9,  8, 10, 12),
    # shield boss (center knob)
    ("d",   12, 34,  8,  6),
    ("m",   13, 35,  6,  4),
    # shield point bottom
    ("d",   10, 36, 12, 10),
    ("m",   12, 37,  8,  8),
    ("d",   14, 44,  4,  6),
    # shield cross divider
    ("d",    6, 20, 22,  3),
    ("acc",  7, 21, 20,  1),
    # shield boss central gem
    ("acc", 14, 36,  4,  4),
    # shield highlight stripe
    ("acc",  9,  8,  4, 10),

    # ---- SWORD (right side) ----
    # blade
    ("l",   38,  5,  6, 34),
    ("m",   40,  5,  4, 34),
    ("acc", 38,  6,  2, 30),
    # blade edge highlight
    ("l",   38,  5,  2,  4),
    # crossguard
    ("d",   34, 38, 16,  5),
    ("m",   35, 39, 14,  3),
    ("acc", 38, 39,  8,  3),
    # grip
    ("d",   39, 43,  6, 12),
    ("m",   40, 44,  4, 10),
    ("acc", 40, 43,  4,  2),
    # pommel
    ("d",   38, 55,  8,  4),
    ("m",   39, 55,  6,  4),
    ("acc", 40, 56,  4,  2),
    # blade tip
    ("l",   39, 39,  2,  2),
    ("m",   40, 37,  2,  4),
]

# GREAT AXE – large two-handed axe
ITEMS["great_axe"] = [
    # handle (center shaft)
    ("d",   29,  6,  6, 52),
    ("m",   30,  6,  4, 52),
    ("acc", 30,  6,  2, 48),
    # handle grip wrapping
    ("d",   28, 24, 8,  4),
    ("acc", 29, 25, 6,  2),
    ("d",   28, 34, 8,  4),
    ("acc", 29, 35, 6,  2),
    ("d",   28, 44, 8,  4),
    ("acc", 29, 45, 6,  2),

    # left axe head (main)
    ("d",    6,  6, 24, 26),
    ("m",    8,  6, 20, 26),
    ("l",   10,  7, 12, 14),
    # left blade edge (sharp)
    ("l",    6, 10,  4, 16),
    ("acc",  6, 12,  2, 10),
    # left head concave curve sim
    ("d",   22, 18,  8, 12),
    # left head lower horn
    ("m",    8, 28, 16,  8),
    ("d",    6, 28, 20,  4),

    # right axe head (secondary, smaller)
    ("d",   34,  8, 16, 18),
    ("m",   34,  8, 14, 18),
    ("l",   35,  9,  8,  8),
    # right blade edge
    ("l",   48, 10,  4, 12),
    ("acc", 50, 12,  2,  8),
    # right head lower horn
    ("m",   34, 24, 12,  6),
    ("d",   34, 26, 14,  4),

    # axe head gem/rune inset
    ("acc", 10, 10,  6,  6),
    ("l",   11, 11,  4,  4),
]

# STAFF – mage staff with orb
ITEMS["staff"] = [
    # shaft
    ("d",   29,  6,  6, 52),
    ("m",   30,  7,  4, 50),
    ("acc", 30,  7,  2, 48),
    # shaft bands
    ("d",   28, 32,  8,  3),
    ("acc", 29, 33,  6,  1),
    ("d",   28, 42,  8,  3),
    ("acc", 29, 43,  6,  1),

    # orb housing top
    ("d",   18,  6, 28,  6),
    ("m",   20,  6, 24,  6),
    # orb housing sides
    ("d",   16, 10,  6, 14),
    ("d",   42, 10,  6, 14),
    # orb glow center
    ("l",   20, 10, 24, 14),
    ("acc", 22, 12, 20, 10),
    # orb highlight
    ("l",   22, 12,  6,  4),
    # orb housing bottom connectors
    ("d",   22, 24,  6,  4),
    ("d",   36, 24,  6,  4),
    # top finial
    ("acc", 28,  4,  8,  4),
    ("l",   30,  4,  4,  2),
    # base ferrule
    ("acc", 28, 56,  8,  2),
    ("d",   26, 54,  12, 4),
    ("m",   27, 55,  10, 3),
]

# SPELLBOOK – arcane tome
ITEMS["spellbook"] = [
    # spine (left binding)
    ("d",    8,  7, 10, 48),
    ("m",   10,  7,  6, 48),
    ("acc", 10,  8,  2, 46),

    # left page stack
    ("l",   18,  7, 16, 48),
    ("m",   18,  7, 14, 48),
    # left page lines
    ("d",   20, 14, 10,  2),
    ("d",   20, 20, 10,  2),
    ("d",   20, 26, 10,  2),
    ("d",   20, 32, 10,  2),
    ("d",   20, 38, 10,  2),
    ("d",   20, 44, 10,  2),

    # center page turn / fold
    ("d",   34,  7,  4, 48),
    ("m",   35,  7,  2, 48),

    # right page stack
    ("l",   38,  7, 16, 48),
    ("m",   38,  7, 14, 48),
    # arcane rune on right page (cross/star)
    ("acc", 42, 16,  8,  2),
    ("acc", 45, 13,  2,  8),
    ("acc", 43, 14,  4,  4),
    # secondary rune dots
    ("acc", 40, 28,  2,  2),
    ("acc", 46, 28,  2,  2),
    ("acc", 43, 34,  2,  2),
    ("acc", 40, 40,  4,  4),

    # cover corners
    ("d",    8,  7,  4,  4),
    ("d",    8, 51,  4,  4),
    ("d",   52,  7,  4,  4),
    ("d",   52, 51,  4,  4),
    # clasp
    ("acc", 54, 26,  4, 12),
    ("l",   55, 28,  2,  8),
]

# SCEPTER – royal scepter
ITEMS["scepter"] = [
    # shaft
    ("d",   29,  6,  6, 52),
    ("m",   30,  7,  4, 50),
    ("acc", 30,  7,  2, 50),
    # shaft rings
    ("d",   27, 28, 10,  4),
    ("m",   28, 29,  8,  2),
    ("acc", 29, 29,  6,  2),

    # head orb base
    ("d",   22, 10, 20,  4),
    ("m",   23, 10, 18,  4),
    # head crown
    ("d",   18, 14, 28, 16),
    ("m",   20, 14, 24, 16),
    ("l",   22, 15, 20, 10),
    # crown jewels left
    ("acc", 18, 14,  4,  8),
    ("l",   19, 15,  2,  6),
    # crown jewels right
    ("acc", 42, 14,  4,  8),
    ("l",   43, 15,  2,  6),
    # crown jewels top
    ("acc", 26, 10,  4,  8),
    ("l",   27, 11,  2,  6),
    ("acc", 34, 10,  4,  8),
    ("l",   35, 11,  2,  6),
    # center fleur
    ("acc", 28, 10,  8,  4),
    ("l",   30, 10,  4,  4),
    # crown gem
    ("acc", 26, 18, 12,  8),
    ("l",   28, 19,  8,  6),
    ("d",   30, 20,  4,  4),

    # base pommel
    ("d",   24, 54, 16,  4),
    ("m",   26, 54, 12,  4),
    ("acc", 28, 55,  8,  2),
]

# HOLY TOME – sacred book with cross
ITEMS["holy_tome"] = [
    # thick cover (hard cover)
    ("d",    6,  6, 52,  4),
    ("d",    6, 54, 52,  4),
    ("d",    6,  6,  8, 52),
    ("d",   50,  6,  8, 52),
    ("m",    8,  8, 48, 48),
    # cover texture
    ("l",   10, 10, 24, 16),
    ("l",   34, 10, 16, 12),
    # spine pages
    ("d",   10, 10, 44, 44),
    ("l",   11, 11, 42, 42),
    # page contents area
    ("m",   12, 12, 40, 40),
    # large cross symbol
    ("acc", 28,  8,  8, 48),
    ("l",   30,  8,  4, 48),
    ("acc",  6, 26, 52,  8),
    ("l",    6, 28, 52,  4),
    # cross center gem
    ("l",   26, 24, 12, 12),
    ("acc", 28, 26,  8,  8),
    ("d",   30, 28,  4,  4),
    # page corners gold leaf
    ("acc", 10, 10,  6,  6),
    ("acc", 10, 48,  6,  6),
    ("acc", 48, 10,  6,  6),
    ("acc", 48, 48,  6,  6),
    # clasp
    ("d",   52, 28,  4,  8),
    ("acc", 53, 29,  2,  6),
]

# WEAPON (generic – short straight sword)
ITEMS["weapon"] = [
    # blade
    ("l",   27,  6, 10, 36),
    ("m",   29,  6,  6, 36),
    ("acc", 27,  7,  2, 34),
    # blade tip
    ("l",   29,  4,  6,  4),
    ("m",   30,  4,  4,  4),
    # fuller (center groove)
    ("d",   31,  8,  2, 30),
    # crossguard
    ("d",   17, 41, 30,  5),
    ("m",   18, 42, 28,  3),
    ("acc", 28, 41,  8,  5),
    # grip
    ("d",   28, 46,  8, 12),
    ("m",   29, 46,  6, 12),
    ("d",   28, 52,  8,  2),
    # grip wrapping accent
    ("acc", 29, 48,  6,  2),
    ("acc", 29, 52,  6,  2),
    # pommel
    ("d",   26, 58, 12,  4),
    ("m",   27, 58, 10,  4),
    ("acc", 29, 59,  6,  2),
]

# UNKNOWN (mystery item – question mark block)
ITEMS["unknown"] = [
    # outer block
    ("d",    8,  8, 48, 48),
    ("m",   10,  9, 44, 46),
    ("l",   10, 10, 24, 16),
    # inner fill
    ("bg2", 12, 12, 40, 40),
    # question mark body (vertical stroke)
    ("m",   26, 18, 12,  8),
    ("l",   28, 18,  8,  8),
    ("m",   26, 24, 14,  8),
    ("l",   28, 24, 10,  8),
    # question mark curve
    ("m",   22, 14, 20,  6),
    ("l",   24, 14, 16,  6),
    ("m",   18, 18,  8, 10),
    ("l",   18, 18,  6, 10),
    ("m",   36, 18,  8, 10),
    # question mark middle void
    ("bg2", 26, 22, 12,  8),
    # question mark stem break
    ("bg2", 26, 30, 14, 6),
    # dot
    ("acc", 28, 40,  8,  8),
    ("l",   30, 41,  4,  4),
    # corner chips
    ("acc", 12, 12,  4,  4),
    ("acc", 48, 12,  4,  4),
    ("acc", 12, 48,  4,  4),
    ("acc", 48, 48,  4,  4),
]

# ---------------------------------------------------------------------------
# Badge thumbnails (top-left inset, 18×18 backdrop + 14×14 symbol at offset 2,2)
# ---------------------------------------------------------------------------
# These are small iconic shapes from the original sprite, scaled to fill 14×14,
# placed in the top-left of the icon at translate(4,4).

def badge_svg(theme: str) -> str:
    """Return SVG group for the small dungeon badge in top-left corner."""
    if theme == "ancient_catacomb":
        # Catacomb arch / skull window
        return (
            '<g transform="translate(4,4)">'
            '<rect x="0" y="1" width="14" height="12" fill="#3f364d"/>'
            '<rect x="2" y="3" width="10" height="8" fill="#7f7ca8"/>'
            '<rect x="4" y="5" width="6" height="4" fill="#c9c7dd"/>'
            '<rect x="5" y="6" width="2" height="2" fill="#2a2433"/>'
            '<rect x="8" y="6" width="2" height="2" fill="#2a2433"/>'
            '<rect x="5" y="9" width="4" height="1" fill="#2a2433"/>'
            '<rect x="0" y="13" width="14" height="1" fill="#2a2433"/>'
            '</g>'
        )
    elif theme == "thorned_hollow":
        # Thorn branch
        return (
            '<g transform="translate(4,4)">'
            '<rect x="5" y="0" width="4" height="14" fill="#2e5130"/>'
            '<rect x="2" y="2" width="4" height="4" fill="#6a9d58"/>'
            '<rect x="8" y="2" width="4" height="4" fill="#79ad61"/>'
            '<rect x="0" y="5" width="5" height="4" fill="#95c36d"/>'
            '<rect x="9" y="5" width="5" height="4" fill="#95c36d"/>'
            '<rect x="1" y="9" width="4" height="4" fill="#587f44"/>'
            '<rect x="9" y="9" width="4" height="4" fill="#587f44"/>'
            '<rect x="0" y="7" width="2" height="1" fill="#d2d88e"/>'
            '<rect x="12" y="7" width="2" height="1" fill="#d2d88e"/>'
            '</g>'
        )
    elif theme == "sunscar_warvault":
        # Sun cross / warvault symbol
        return (
            '<g transform="translate(4,4)">'
            '<rect x="4" y="0" width="6" height="5" fill="#f5d565"/>'
            '<rect x="3" y="4" width="8" height="3" fill="#d75333"/>'
            '<rect x="5" y="3" width="4" height="10" fill="#f0b54d"/>'
            '<rect x="4" y="8" width="6" height="6" fill="#a33a24"/>'
            '<rect x="2" y="9" width="3" height="3" fill="#f5d565"/>'
            '<rect x="9" y="9" width="3" height="3" fill="#f5d565"/>'
            '<rect x="1" y="5" width="3" height="3" fill="#f5d565"/>'
            '<rect x="10" y="5" width="3" height="3" fill="#f5d565"/>'
            '</g>'
        )
    elif theme == "obsidian_spire":
        # Spire / tower shape
        return (
            '<g transform="translate(4,4)">'
            '<rect x="5" y="0" width="4" height="3" fill="#b8b0df"/>'
            '<rect x="4" y="2" width="6" height="4" fill="#6d6498"/>'
            '<rect x="3" y="6" width="8" height="4" fill="#433a61"/>'
            '<rect x="2" y="10" width="10" height="4" fill="#272235"/>'
            '<rect x="1" y="13" width="12" height="1" fill="#17141f"/>'
            '<rect x="4" y="6" width="2" height="3" fill="#d7d2f1"/>'
            '<rect x="8" y="6" width="2" height="3" fill="#d7d2f1"/>'
            '</g>'
        )
    elif theme == "starter":
        # Starter stone block
        return (
            '<g transform="translate(4,4)">'
            '<rect x="1" y="1" width="12" height="12" fill="#b7ac8c"/>'
            '<rect x="3" y="3" width="8" height="8" fill="#efe5c8"/>'
            '<rect x="4" y="4" width="6" height="6" fill="#7c735d"/>'
            '<rect x="2" y="13" width="10" height="1" fill="#635b48"/>'
            '</g>'
        )
    else:  # unknown
        return (
            '<g transform="translate(4,4)">'
            '<rect x="0" y="0" width="14" height="14" fill="#6d6758"/>'
            '<rect x="2" y="2" width="10" height="10" fill="#d9d0b7"/>'
            '<rect x="5" y="4" width="4" height="4" fill="#443f3a"/>'
            '<rect x="6" y="9" width="2" height="2" fill="#443f3a"/>'
            '</g>'
        )


# ---------------------------------------------------------------------------
# SVG builder
# ---------------------------------------------------------------------------

def build_svg(theme_name: str, item_name: str) -> str:
    t = THEMES[theme_name]
    rects = ITEMS[item_name]

    def color(key: str) -> str:
        if key.startswith("#"):
            return key
        return t[key]

    # Background
    lines = [
        '<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 64 64">',
        f'  <rect x="0" y="0" width="64" height="64" fill="{t["bg"]}"/>',
        f'  <rect x="4" y="4" width="56" height="56" fill="{t["bg2"]}" opacity="0.45"/>',
        # outer frame
        f'  <rect x="2" y="2" width="60" height="60" fill="none" stroke="{t["frame"]}" stroke-opacity="0.65" stroke-width="2"/>',
        # inner frame
        f'  <rect x="6" y="6" width="52" height="52" fill="none" stroke="{t["frame2"]}" stroke-opacity="0.30" stroke-width="1"/>',
    ]

    # item rects
    for entry in rects:
        c, x, y, w, h = entry
        lines.append(f'  <rect x="{x}" y="{y}" width="{w}" height="{h}" fill="{color(c)}"/>')

    # badge in top-left, drawn on top
    lines.append(f"  {badge_svg(theme_name)}")

    # badge backdrop
    lines.insert(5, f'  <rect x="3" y="3" width="20" height="20" fill="{t["bg"]}" opacity="0.72"/>')

    lines.append("</svg>")
    return "\n".join(lines)


# ---------------------------------------------------------------------------
# Main
# ---------------------------------------------------------------------------

def main() -> None:
    script_dir = os.path.dirname(os.path.abspath(__file__))
    repo_root = os.path.dirname(script_dir)
    out_base = os.path.join(repo_root, OUT_BASE)

    generated = 0
    for theme_name in THEMES:
        theme_dir = os.path.join(out_base, theme_name)
        os.makedirs(theme_dir, exist_ok=True)
        for item_name in ITEMS:
            svg = build_svg(theme_name, item_name)
            path = os.path.join(theme_dir, f"{item_name}.svg")
            with open(path, "w", encoding="utf-8") as fh:
                fh.write(svg)
            generated += 1

    print(f"Generated {generated} SVG icons under {OUT_BASE}/")


if __name__ == "__main__":
    main()
