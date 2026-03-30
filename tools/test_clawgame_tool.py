from __future__ import annotations

import json
import tempfile
import unittest
from pathlib import Path

from tools import clawgame_tool


class ClawGameToolTests(unittest.TestCase):
    def test_solve_challenge_prompt(self) -> None:
        prompt = "Cipher check: ember=12, frost=8, moss=5, factor=3. Reply with digits only."
        self.assertEqual(clawgame_tool.solve_challenge_prompt(prompt), "45")

    def test_coerce_scalar(self) -> None:
        self.assertIs(clawgame_tool.coerce_scalar("true"), True)
        self.assertIs(clawgame_tool.coerce_scalar("false"), False)
        self.assertEqual(clawgame_tool.coerce_scalar("12"), 12)
        self.assertEqual(clawgame_tool.coerce_scalar("3.5"), 3.5)
        self.assertEqual(clawgame_tool.coerce_scalar("mage"), "mage")

    def test_parse_key_value_pairs(self) -> None:
        parsed = clawgame_tool.parse_key_value_pairs(["region_id=greenfield_village", "difficulty=hard", "count=2"])
        self.assertEqual(
            parsed,
            {
                "region_id": "greenfield_village",
                "difficulty": "hard",
                "count": 2,
            },
        )

    def test_state_round_trip(self) -> None:
        with tempfile.TemporaryDirectory() as tmp_dir:
            path = Path(tmp_dir) / "state.json"
            payload = {"bot_name": "bot-one", "pending_claim_run_ids": ["run_1"]}
            clawgame_tool.save_state(path, payload)
            self.assertEqual(clawgame_tool.load_state(path), payload)
            self.assertEqual(json.loads(path.read_text(encoding="utf-8"))["bot_name"], "bot-one")


if __name__ == "__main__":
    unittest.main()
