#!/usr/bin/env python3
from __future__ import annotations

import argparse
import json
import re
import sys
import urllib.error
import urllib.parse
import urllib.request
from dataclasses import dataclass
from datetime import datetime, timedelta, timezone
from pathlib import Path
from typing import Any, Callable

DEFAULT_API_BASE = "http://localhost:8080/api/v1"
DEFAULT_OBSERVER_ORIGIN = "http://localhost:4000"
DEFAULT_STATE_FILE = ".openclaw/clawgame-state.json"
CHALLENGE_PATTERN = re.compile(r"ember=(\d+).+frost=(\d+).+moss=(\d+).+factor=(\d+)")


@dataclass
class CommandResult:
    command: str
    data: Any
    request_id: str | None = None


@dataclass
class RuntimeContext:
    api_base: str
    observer_origin: str
    state_file: Path
    pretty: bool
    timeout_seconds: float
    explicit_access_token: str | None
    state: dict[str, Any]


class CommandFailure(Exception):
    def __init__(self, exit_code: int, payload: dict[str, Any]):
        super().__init__(payload.get("error", {}).get("message", "command failed"))
        self.exit_code = exit_code
        self.payload = payload


class APIClient:
    def __init__(self, api_base: str, observer_origin: str, timeout_seconds: float):
        self.api_base = api_base.rstrip("/")
        self.observer_origin = observer_origin.rstrip("/")
        self.timeout_seconds = timeout_seconds

    def request(
        self,
        command: str,
        method: str,
        path: str,
        *,
        body: Any | None = None,
        query: dict[str, Any] | None = None,
        token: str | None = None,
        base: str = "api",
    ) -> dict[str, Any]:
        url = self._build_url(path, query=query, base=base)
        data_bytes: bytes | None = None
        if method.upper() != "GET":
            payload = {} if body is None else body
            data_bytes = json.dumps(payload).encode("utf-8")

        request = urllib.request.Request(url=url, data=data_bytes, method=method.upper())
        request.add_header("Accept", "application/json")
        if data_bytes is not None:
            request.add_header("Content-Type", "application/json")
        if token:
            request.add_header("Authorization", f"Bearer {token}")

        try:
            with urllib.request.urlopen(request, timeout=self.timeout_seconds) as response:
                raw = response.read().decode("utf-8")
                envelope = json.loads(raw) if raw else {}
        except urllib.error.HTTPError as exc:
            raw = exc.read().decode("utf-8", errors="replace")
            try:
                envelope = json.loads(raw) if raw else {}
            except json.JSONDecodeError:
                envelope = {}
            raise self._remote_error(command, envelope, fallback_message=f"HTTP {exc.code}") from exc
        except urllib.error.URLError as exc:
            raise CommandFailure(
                5,
                {
                    "ok": False,
                    "command": command,
                    "error": {
                        "code": "NETWORK_ERROR",
                        "message": str(exc.reason),
                    },
                },
            ) from exc
        except json.JSONDecodeError as exc:
            raise CommandFailure(
                5,
                {
                    "ok": False,
                    "command": command,
                    "error": {
                        "code": "INVALID_JSON_RESPONSE",
                        "message": f"could not decode JSON response: {exc}",
                    },
                },
            ) from exc

        if isinstance(envelope, dict) and isinstance(envelope.get("error"), dict):
            raise self._remote_error(command, envelope)
        if not isinstance(envelope, dict):
            raise CommandFailure(
                5,
                {
                    "ok": False,
                    "command": command,
                    "error": {
                        "code": "INVALID_ENVELOPE",
                        "message": "response was not a JSON object",
                    },
                },
            )
        return envelope

    def _build_url(self, path: str, *, query: dict[str, Any] | None, base: str) -> str:
        if path.startswith("http://") or path.startswith("https://"):
            base_url = path
        else:
            root = self.api_base if base == "api" else self.observer_origin
            base_url = f"{root}/{path.lstrip('/')}"
        if query:
            cleaned = {key: value for key, value in query.items() if value is not None}
            if cleaned:
                sep = "&" if "?" in base_url else "?"
                base_url = f"{base_url}{sep}{urllib.parse.urlencode(cleaned)}"
        return base_url

    def _remote_error(
        self,
        command: str,
        envelope: dict[str, Any],
        *,
        fallback_message: str | None = None,
    ) -> CommandFailure:
        error = envelope.get("error") if isinstance(envelope, dict) else None
        request_id = envelope.get("request_id") if isinstance(envelope, dict) else None
        code = "API_ERROR"
        message = fallback_message or "remote API returned an error"
        if isinstance(error, dict):
            code = str(error.get("code") or code)
            message = str(error.get("message") or message)
        return CommandFailure(
            3,
            {
                "ok": False,
                "command": command,
                "error": {
                    "code": code,
                    "message": message,
                    "request_id": request_id,
                },
            },
        )


def parse_datetime(value: str | None) -> datetime | None:
    if not value:
        return None
    normalized = value.replace("Z", "+00:00")
    try:
        return datetime.fromisoformat(normalized)
    except ValueError:
        return None


def token_is_usable(expires_at: str | None, *, skew_seconds: int = 30) -> bool:
    if not expires_at:
        return True
    parsed = parse_datetime(expires_at)
    if parsed is None:
        return True
    now = datetime.now(timezone.utc)
    if parsed.tzinfo is None:
        parsed = parsed.replace(tzinfo=timezone.utc)
    return parsed > now + timedelta(seconds=skew_seconds)


def load_state(path: Path) -> dict[str, Any]:
    if not path.exists():
        return {}
    try:
        loaded = json.loads(path.read_text(encoding="utf-8"))
    except json.JSONDecodeError as exc:
        raise CommandFailure(
            4,
            {
                "ok": False,
                "command": "state",
                "error": {
                    "code": "STATE_FILE_INVALID",
                    "message": f"could not parse state file: {exc}",
                },
            },
        ) from exc
    if not isinstance(loaded, dict):
        raise CommandFailure(
            4,
            {
                "ok": False,
                "command": "state",
                "error": {
                    "code": "STATE_FILE_INVALID",
                    "message": "state file must contain a JSON object",
                },
            },
        )
    return loaded


def save_state(path: Path, state: dict[str, Any]) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text(json.dumps(state, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")


def emit_json(payload: dict[str, Any], *, pretty: bool) -> None:
    if pretty:
        print(json.dumps(payload, ensure_ascii=False, indent=2))
    else:
        print(json.dumps(payload, ensure_ascii=False, separators=(",", ":")))


def success_payload(ctx: RuntimeContext, result: CommandResult) -> dict[str, Any]:
    meta: dict[str, Any] = {
        "api_base_url": ctx.api_base,
        "state_file": str(ctx.state_file),
    }
    if result.request_id:
        meta["request_id"] = result.request_id
    return {
        "ok": True,
        "command": result.command,
        "data": result.data,
        "meta": meta,
    }


def coerce_scalar(value: str) -> Any:
    lowered = value.lower()
    if lowered == "true":
        return True
    if lowered == "false":
        return False
    if lowered == "null":
        return None
    if re.fullmatch(r"-?\d+", value):
        try:
            return int(value)
        except ValueError:
            pass
    if re.fullmatch(r"-?\d+\.\d+", value):
        try:
            return float(value)
        except ValueError:
            pass
    return value


def parse_key_value_pairs(items: list[str]) -> dict[str, Any]:
    result: dict[str, Any] = {}
    for item in items:
        if "=" not in item:
            raise CommandFailure(
                2,
                {
                    "ok": False,
                    "command": "action",
                    "error": {
                        "code": "INVALID_KEY_VALUE",
                        "message": f"expected key=value, got: {item}",
                    },
                },
            )
        key, raw = item.split("=", 1)
        result[key] = coerce_scalar(raw)
    return result


def parse_bool_text(value: str | None, *, default: bool) -> bool:
    if value is None:
        return default
    lowered = value.lower()
    if lowered == "true":
        return True
    if lowered == "false":
        return False
    raise CommandFailure(
        2,
        {
            "ok": False,
            "command": "arguments",
            "error": {
                "code": "INVALID_BOOLEAN",
                "message": f"expected true or false, got: {value}",
            },
        },
    )


def solve_challenge_prompt(prompt: str) -> str:
    match = CHALLENGE_PATTERN.search(prompt)
    if not match:
        raise CommandFailure(
            4,
            {
                "ok": False,
                "command": "auth challenge",
                "error": {
                    "code": "UNSUPPORTED_CHALLENGE_FORMAT",
                    "message": f"unexpected challenge prompt format: {prompt}",
                },
            },
        )
    ember, frost, moss, factor = (int(group) for group in match.groups())
    return str(((ember + frost) - moss) * factor)


def require_text(value: str | None, *, command: str, field: str, fallback: str | None = None) -> str:
    chosen = value or fallback
    if chosen:
        return chosen
    raise CommandFailure(
        2,
        {
            "ok": False,
            "command": command,
            "error": {
                "code": "MISSING_ARGUMENT",
                "message": f"missing required field: {field}",
            },
        },
    )


def update_tokens(state: dict[str, Any], login_data: dict[str, Any]) -> None:
    for key in (
        "access_token",
        "access_token_expires_at",
        "refresh_token",
        "refresh_token_expires_at",
    ):
        if key in login_data:
            state[key] = login_data[key]


def sync_character_summary(state: dict[str, Any], character: dict[str, Any] | None) -> None:
    if not isinstance(character, dict):
        return
    if isinstance(character.get("character_id"), str):
        state["character_id"] = character["character_id"]
    if isinstance(character.get("name"), str):
        state["character_name"] = character["name"]
    if isinstance(character.get("gender"), str):
        state["gender"] = character["gender"]
    if isinstance(character.get("class"), str):
        state["class"] = character["class"]
    if isinstance(character.get("weapon_style"), str):
        state["weapon_style"] = character["weapon_style"]
    region = character.get("location_region_id") or character.get("region_id")
    if isinstance(region, str):
        state["last_region_id"] = region


def sync_me_like_payload(state: dict[str, Any], data: dict[str, Any] | None) -> None:
    if not isinstance(data, dict):
        return
    sync_character_summary(state, data.get("character"))
    if isinstance(data.get("character_region_id"), str):
        state["last_region_id"] = data["character_region_id"]
    dungeon_daily = data.get("dungeon_daily")
    if isinstance(dungeon_daily, dict):
        pending = dungeon_daily.get("pending_claim_run_ids")
        if isinstance(pending, list):
            state["pending_claim_run_ids"] = [item for item in pending if isinstance(item, str)]


def sync_region_detail(state: dict[str, Any], detail: dict[str, Any] | None) -> None:
    if not isinstance(detail, dict):
        return
    region = detail.get("region")
    if isinstance(region, dict):
        region_id = region.get("region_id")
        if isinstance(region_id, str):
            state["last_region_id"] = region_id


def sync_run_view(state: dict[str, Any], run_view: dict[str, Any] | None) -> None:
    if not isinstance(run_view, dict):
        return
    run_id = run_view.get("run_id")
    if isinstance(run_id, str):
        state["last_run_id"] = run_id
        pending = list(state.get("pending_claim_run_ids") or [])
        pending_set = {item for item in pending if isinstance(item, str)}
        if run_view.get("reward_claimed_at"):
            pending_set.discard(run_id)
        elif run_view.get("reward_claimable"):
            pending_set.add(run_id)
        state["pending_claim_run_ids"] = sorted(pending_set)


def sync_action_payload(state: dict[str, Any], data: dict[str, Any] | None) -> None:
    if not isinstance(data, dict):
        return
    action_result = data.get("action_result")
    if isinstance(action_result, dict):
        run_id = action_result.get("run_id")
        if isinstance(run_id, str):
            state["last_run_id"] = run_id
            if action_result.get("action_type") == "claim_dungeon_rewards":
                pending = [item for item in state.get("pending_claim_run_ids", []) if item != run_id]
                state["pending_claim_run_ids"] = pending
    nested_state = data.get("state")
    if isinstance(nested_state, dict):
        sync_me_like_payload(state, nested_state)
        run = nested_state.get("run")
        if isinstance(run, dict):
            sync_run_view(state, run)


def set_last_request_id(state: dict[str, Any], envelope: dict[str, Any]) -> None:
    request_id = envelope.get("request_id")
    if isinstance(request_id, str):
        state["last_request_id"] = request_id


def issue_auth_challenge(client: APIClient, state: dict[str, Any]) -> tuple[dict[str, Any], dict[str, Any]]:
    envelope = client.request("auth challenge", "POST", "/auth/challenge", body={})
    set_last_request_id(state, envelope)
    data = envelope.get("data") or {}
    challenge = data.get("challenge") if isinstance(data, dict) else {}
    if not isinstance(challenge, dict):
        raise CommandFailure(
            5,
            {
                "ok": False,
                "command": "auth challenge",
                "error": {
                    "code": "INVALID_CHALLENGE_PAYLOAD",
                    "message": "challenge response did not contain a challenge object",
                },
            },
        )
    return envelope, challenge


def perform_register(
    client: APIClient,
    state: dict[str, Any],
    *,
    bot_name: str,
    password: str,
) -> dict[str, Any]:
    _, challenge = issue_auth_challenge(client, state)
    payload = {
        "bot_name": bot_name,
        "password": password,
        "challenge_id": challenge.get("challenge_id"),
        "challenge_answer": solve_challenge_prompt(str(challenge.get("prompt_text", ""))),
    }
    envelope = client.request("register", "POST", "/auth/register", body=payload)
    set_last_request_id(state, envelope)
    state["bot_name"] = bot_name
    state["password"] = password
    return envelope


def perform_login(
    client: APIClient,
    state: dict[str, Any],
    *,
    bot_name: str,
    password: str,
) -> dict[str, Any]:
    _, challenge = issue_auth_challenge(client, state)
    payload = {
        "bot_name": bot_name,
        "password": password,
        "challenge_id": challenge.get("challenge_id"),
        "challenge_answer": solve_challenge_prompt(str(challenge.get("prompt_text", ""))),
    }
    envelope = client.request("login", "POST", "/auth/login", body=payload)
    set_last_request_id(state, envelope)
    data = envelope.get("data") if isinstance(envelope, dict) else None
    if isinstance(data, dict):
        update_tokens(state, data)
    state["bot_name"] = bot_name
    state["password"] = password
    return envelope


def perform_refresh(client: APIClient, state: dict[str, Any]) -> dict[str, Any]:
    refresh_token = state.get("refresh_token")
    if not isinstance(refresh_token, str) or not refresh_token:
        raise CommandFailure(
            4,
            {
                "ok": False,
                "command": "refresh",
                "error": {
                    "code": "REFRESH_TOKEN_MISSING",
                    "message": "no refresh token available in state",
                },
            },
        )
    envelope = client.request("refresh", "POST", "/auth/refresh", body={"refresh_token": refresh_token})
    set_last_request_id(state, envelope)
    data = envelope.get("data") if isinstance(envelope, dict) else None
    if isinstance(data, dict):
        update_tokens(state, data)
    return envelope


def ensure_access_token(ctx: RuntimeContext, client: APIClient) -> str:
    if ctx.explicit_access_token:
        return ctx.explicit_access_token
    token = ctx.state.get("access_token")
    expires = ctx.state.get("access_token_expires_at")
    if isinstance(token, str) and token and token_is_usable(expires if isinstance(expires, str) else None):
        return token

    refresh_token = ctx.state.get("refresh_token")
    refresh_expires = ctx.state.get("refresh_token_expires_at")
    if isinstance(refresh_token, str) and refresh_token and token_is_usable(refresh_expires if isinstance(refresh_expires, str) else None):
        perform_refresh(client, ctx.state)
        token = ctx.state.get("access_token")
        if isinstance(token, str) and token:
            save_state(ctx.state_file, ctx.state)
            return token

    bot_name = ctx.state.get("bot_name")
    password = ctx.state.get("password")
    if isinstance(bot_name, str) and bot_name and isinstance(password, str) and password:
        perform_login(client, ctx.state, bot_name=bot_name, password=password)
        token = ctx.state.get("access_token")
        if isinstance(token, str) and token:
            save_state(ctx.state_file, ctx.state)
            return token

    raise CommandFailure(
        4,
        {
            "ok": False,
            "command": "auth",
            "error": {
                "code": "AUTH_STATE_MISSING",
                "message": "no usable access path found; run bootstrap or login first",
            },
        },
    )


def authenticated_request(
    ctx: RuntimeContext,
    client: APIClient,
    command: str,
    method: str,
    path: str,
    *,
    body: Any | None = None,
    query: dict[str, Any] | None = None,
) -> dict[str, Any]:
    token = ensure_access_token(ctx, client)
    try:
        envelope = client.request(command, method, path, body=body, query=query, token=token)
    except CommandFailure as exc:
        code = exc.payload.get("error", {}).get("code")
        if code in {"AUTH_TOKEN_EXPIRED", "AUTH_REQUIRED"} and not ctx.explicit_access_token:
            ctx.state.pop("access_token", None)
            ctx.state.pop("access_token_expires_at", None)
            token = ensure_access_token(ctx, client)
            envelope = client.request(command, method, path, body=body, query=query, token=token)
        else:
            raise
    set_last_request_id(ctx.state, envelope)
    return envelope


def extract_data(envelope: dict[str, Any]) -> dict[str, Any] | None:
    data = envelope.get("data")
    if isinstance(data, dict):
        return data
    return data if data is not None else None


def fetch_region_detail(
    ctx: RuntimeContext,
    client: APIClient,
    region_id: str | None,
    *,
    command: str,
) -> tuple[dict[str, Any] | None, str | None]:
    if not isinstance(region_id, str) or not region_id:
        return None, None
    envelope = client.request(command, "GET", f"/regions/{region_id}")
    set_last_request_id(ctx.state, envelope)
    data = extract_data(envelope)
    sync_region_detail(ctx.state, data if isinstance(data, dict) else None)
    return data, envelope.get("request_id")


def cmd_register(args: argparse.Namespace, ctx: RuntimeContext, client: APIClient) -> CommandResult:
    bot_name = require_text(args.bot_name, command="register", field="bot_name", fallback=ctx.state.get("bot_name"))
    password = require_text(args.password, command="register", field="password", fallback=ctx.state.get("password"))
    envelope = perform_register(client, ctx.state, bot_name=bot_name, password=password)
    save_state(ctx.state_file, ctx.state)
    return CommandResult("register", extract_data(envelope), envelope.get("request_id"))


def cmd_login(args: argparse.Namespace, ctx: RuntimeContext, client: APIClient) -> CommandResult:
    bot_name = require_text(args.bot_name, command="login", field="bot_name", fallback=ctx.state.get("bot_name"))
    password = require_text(args.password, command="login", field="password", fallback=ctx.state.get("password"))
    envelope = perform_login(client, ctx.state, bot_name=bot_name, password=password)
    save_state(ctx.state_file, ctx.state)
    return CommandResult("login", extract_data(envelope), envelope.get("request_id"))


def cmd_refresh(args: argparse.Namespace, ctx: RuntimeContext, client: APIClient) -> CommandResult:
    envelope = perform_refresh(client, ctx.state)
    save_state(ctx.state_file, ctx.state)
    return CommandResult("refresh", extract_data(envelope), envelope.get("request_id"))


def cmd_bootstrap(args: argparse.Namespace, ctx: RuntimeContext, client: APIClient) -> CommandResult:
    bot_name = args.bot_name or ctx.state.get("bot_name")
    password = args.password or ctx.state.get("password")
    if bot_name:
        ctx.state["bot_name"] = bot_name
    if password:
        ctx.state["password"] = password
    if args.character_name:
        ctx.state["character_name"] = args.character_name
    if args.gender:
        ctx.state["gender"] = args.gender
    if args.character_class:
        ctx.state["class"] = args.character_class
    if args.weapon_style:
        ctx.state["weapon_style"] = args.weapon_style

    login_ready = isinstance(bot_name, str) and bot_name and isinstance(password, str) and password

    if not ctx.explicit_access_token:
        token = ctx.state.get("access_token")
        expires = ctx.state.get("access_token_expires_at")
        if not (isinstance(token, str) and token and token_is_usable(expires if isinstance(expires, str) else None)):
            refreshed = False
            refresh_token = ctx.state.get("refresh_token")
            if isinstance(refresh_token, str) and refresh_token:
                try:
                    perform_refresh(client, ctx.state)
                    refreshed = True
                except CommandFailure:
                    refreshed = False
            if not refreshed and login_ready:
                try:
                    perform_login(client, ctx.state, bot_name=bot_name, password=password)
                except CommandFailure as exc:
                    register_if_needed = parse_bool_text(args.register_if_needed, default=True)
                    if exc.payload.get("error", {}).get("code") == "AUTH_INVALID_CREDENTIALS" and register_if_needed:
                        try:
                            perform_register(client, ctx.state, bot_name=bot_name, password=password)
                        except CommandFailure as register_exc:
                            if register_exc.payload.get("error", {}).get("code") != "ACCOUNT_BOT_NAME_TAKEN":
                                raise
                        perform_login(client, ctx.state, bot_name=bot_name, password=password)
                    else:
                        raise

    me_envelope = authenticated_request(ctx, client, "me", "GET", "/me")
    me_data = extract_data(me_envelope)
    sync_me_like_payload(ctx.state, me_data)

    character_missing = isinstance(me_data, dict) and me_data.get("character") is None
    if character_missing:
        character_name = args.character_name or ctx.state.get("character_name")
        gender = args.gender or ctx.state.get("gender")
        character_class = args.character_class or ctx.state.get("class")
        weapon_style = args.weapon_style or ctx.state.get("weapon_style")
        if character_name and gender and character_class and weapon_style:
            create_payload = {
                "name": character_name,
                "gender": gender,
                "class": character_class,
                "weapon_style": weapon_style,
            }
            create_envelope = authenticated_request(ctx, client, "characters create", "POST", "/characters", body=create_payload)
            sync_me_like_payload(ctx.state, extract_data(create_envelope))
            me_envelope = create_envelope
            me_data = extract_data(me_envelope)

    planner_envelope = authenticated_request(ctx, client, "planner", "GET", "/me/planner")
    planner_data = extract_data(planner_envelope)
    sync_me_like_payload(ctx.state, planner_data if isinstance(planner_data, dict) else None)
    region_data, region_request_id = fetch_region_detail(
        ctx,
        client,
        ctx.state.get("last_region_id"),
        command="bootstrap region",
    )
    save_state(ctx.state_file, ctx.state)

    return CommandResult(
        "bootstrap",
        {
            "me": me_data,
            "planner": planner_data,
            "region": region_data,
            "state_summary": {
                "bot_name": ctx.state.get("bot_name"),
                "character_name": ctx.state.get("character_name"),
                "gender": ctx.state.get("gender"),
                "character_id": ctx.state.get("character_id"),
                "last_region_id": ctx.state.get("last_region_id"),
                "pending_claim_run_ids": ctx.state.get("pending_claim_run_ids", []),
            },
        },
        region_request_id or planner_envelope.get("request_id") or me_envelope.get("request_id"),
    )


def cmd_me(args: argparse.Namespace, ctx: RuntimeContext, client: APIClient) -> CommandResult:
    envelope = authenticated_request(ctx, client, "me", "GET", "/me")
    data = extract_data(envelope)
    sync_me_like_payload(ctx.state, data)
    save_state(ctx.state_file, ctx.state)
    return CommandResult("me", data, envelope.get("request_id"))


def cmd_planner(args: argparse.Namespace, ctx: RuntimeContext, client: APIClient) -> CommandResult:
    query = {"region_id": args.region_id} if args.region_id else None
    envelope = authenticated_request(ctx, client, "planner", "GET", "/me/planner", query=query)
    data = extract_data(envelope)
    sync_me_like_payload(ctx.state, data)
    save_state(ctx.state_file, ctx.state)
    return CommandResult("planner", data, envelope.get("request_id"))


def cmd_state(args: argparse.Namespace, ctx: RuntimeContext, client: APIClient) -> CommandResult:
    envelope = authenticated_request(ctx, client, "state", "GET", "/me/state")
    data = extract_data(envelope)
    sync_me_like_payload(ctx.state, data)
    save_state(ctx.state_file, ctx.state)
    return CommandResult("state", data, envelope.get("request_id"))


def cmd_actions(args: argparse.Namespace, ctx: RuntimeContext, client: APIClient) -> CommandResult:
    envelope = authenticated_request(ctx, client, "actions", "GET", "/me/actions")
    save_state(ctx.state_file, ctx.state)
    return CommandResult("actions", extract_data(envelope), envelope.get("request_id"))


def cmd_field(args: argparse.Namespace, ctx: RuntimeContext, client: APIClient) -> CommandResult:
    envelope = authenticated_request(
        ctx,
        client,
        f"field {args.approach}",
        "POST",
        "/me/field-encounter",
        body={"approach": args.approach},
    )
    data = extract_data(envelope)
    sync_action_payload(ctx.state, data if isinstance(data, dict) else None)
    save_state(ctx.state_file, ctx.state)
    return CommandResult(f"field {args.approach}", data, envelope.get("request_id"))


def cmd_regions_list(args: argparse.Namespace, ctx: RuntimeContext, client: APIClient) -> CommandResult:
    envelope = client.request("regions list", "GET", "/world/regions")
    set_last_request_id(ctx.state, envelope)
    return CommandResult("regions list", extract_data(envelope), envelope.get("request_id"))


def cmd_regions_show(args: argparse.Namespace, ctx: RuntimeContext, client: APIClient) -> CommandResult:
    envelope = client.request("regions show", "GET", f"/regions/{args.region_id}")
    set_last_request_id(ctx.state, envelope)
    data = extract_data(envelope)
    sync_region_detail(ctx.state, data if isinstance(data, dict) else None)
    save_state(ctx.state_file, ctx.state)
    return CommandResult("regions show", data, envelope.get("request_id"))


def cmd_travel(args: argparse.Namespace, ctx: RuntimeContext, client: APIClient) -> CommandResult:
    envelope = authenticated_request(ctx, client, "travel", "POST", "/me/travel", body={"region_id": args.region_id})
    data = extract_data(envelope)
    sync_action_payload(ctx.state, data if isinstance(data, dict) else None)
    planner_envelope = authenticated_request(ctx, client, "travel planner", "GET", "/me/planner", query={"region_id": args.region_id})
    planner_data = extract_data(planner_envelope)
    sync_me_like_payload(ctx.state, planner_data if isinstance(planner_data, dict) else None)
    region_data, region_request_id = fetch_region_detail(ctx, client, args.region_id, command="travel region")
    save_state(ctx.state_file, ctx.state)
    return CommandResult(
        "travel",
        {
            "travel": data,
            "planner": planner_data,
            "region": region_data,
        },
        region_request_id or planner_envelope.get("request_id") or envelope.get("request_id"),
    )


def cmd_quests_list(args: argparse.Namespace, ctx: RuntimeContext, client: APIClient) -> CommandResult:
    envelope = authenticated_request(ctx, client, "quests list", "GET", "/me/quests")
    save_state(ctx.state_file, ctx.state)
    return CommandResult("quests list", extract_data(envelope), envelope.get("request_id"))


def cmd_quests_show(args: argparse.Namespace, ctx: RuntimeContext, client: APIClient) -> CommandResult:
    envelope = authenticated_request(ctx, client, "quests show", "GET", f"/me/quests/{args.quest_id}")
    save_state(ctx.state_file, ctx.state)
    return CommandResult("quests show", extract_data(envelope), envelope.get("request_id"))


def cmd_quests_accept(args: argparse.Namespace, ctx: RuntimeContext, client: APIClient) -> CommandResult:
    envelope = authenticated_request(ctx, client, "quests accept", "POST", f"/me/quests/{args.quest_id}/accept")
    data = extract_data(envelope)
    sync_action_payload(ctx.state, data if isinstance(data, dict) else None)
    save_state(ctx.state_file, ctx.state)
    return CommandResult("quests accept", data, envelope.get("request_id"))


def cmd_quests_choice(args: argparse.Namespace, ctx: RuntimeContext, client: APIClient) -> CommandResult:
    envelope = authenticated_request(
        ctx,
        client,
        "quests choice",
        "POST",
        f"/me/quests/{args.quest_id}/choice",
        body={"choice_key": args.choice_key},
    )
    data = extract_data(envelope)
    sync_action_payload(ctx.state, data if isinstance(data, dict) else None)
    save_state(ctx.state_file, ctx.state)
    return CommandResult("quests choice", data, envelope.get("request_id"))


def cmd_quests_interact(args: argparse.Namespace, ctx: RuntimeContext, client: APIClient) -> CommandResult:
    envelope = authenticated_request(
        ctx,
        client,
        "quests interact",
        "POST",
        f"/me/quests/{args.quest_id}/interact",
        body={"interaction": args.interaction},
    )
    data = extract_data(envelope)
    sync_action_payload(ctx.state, data if isinstance(data, dict) else None)
    save_state(ctx.state_file, ctx.state)
    return CommandResult("quests interact", data, envelope.get("request_id"))


def cmd_quests_submit(args: argparse.Namespace, ctx: RuntimeContext, client: APIClient) -> CommandResult:
    envelope = authenticated_request(ctx, client, "quests submit", "POST", f"/me/quests/{args.quest_id}/submit")
    data = extract_data(envelope)
    sync_action_payload(ctx.state, data if isinstance(data, dict) else None)
    save_state(ctx.state_file, ctx.state)
    return CommandResult("quests submit", data, envelope.get("request_id"))


def cmd_quests_reroll(args: argparse.Namespace, ctx: RuntimeContext, client: APIClient) -> CommandResult:
    confirm_cost = parse_bool_text(args.confirm_cost, default=True)
    envelope = authenticated_request(ctx, client, "quests reroll", "POST", "/me/quests/reroll", body={"confirm_cost": confirm_cost})
    data = extract_data(envelope)
    sync_action_payload(ctx.state, data if isinstance(data, dict) else None)
    save_state(ctx.state_file, ctx.state)
    return CommandResult("quests reroll", data, envelope.get("request_id"))


def cmd_inventory(args: argparse.Namespace, ctx: RuntimeContext, client: APIClient) -> CommandResult:
    envelope = authenticated_request(ctx, client, "inventory", "GET", "/me/inventory")
    save_state(ctx.state_file, ctx.state)
    return CommandResult("inventory", extract_data(envelope), envelope.get("request_id"))


def cmd_equipment_equip(args: argparse.Namespace, ctx: RuntimeContext, client: APIClient) -> CommandResult:
    envelope = authenticated_request(ctx, client, "equipment equip", "POST", "/me/equipment/equip", body={"item_id": args.item_id})
    data = extract_data(envelope)
    sync_action_payload(ctx.state, data if isinstance(data, dict) else None)
    save_state(ctx.state_file, ctx.state)
    return CommandResult("equipment equip", data, envelope.get("request_id"))


def cmd_equipment_unequip(args: argparse.Namespace, ctx: RuntimeContext, client: APIClient) -> CommandResult:
    envelope = authenticated_request(ctx, client, "equipment unequip", "POST", "/me/equipment/unequip", body={"slot": args.slot})
    data = extract_data(envelope)
    sync_action_payload(ctx.state, data if isinstance(data, dict) else None)
    save_state(ctx.state_file, ctx.state)
    return CommandResult("equipment unequip", data, envelope.get("request_id"))


def cmd_buildings_show(args: argparse.Namespace, ctx: RuntimeContext, client: APIClient) -> CommandResult:
    envelope = client.request("buildings show", "GET", f"/buildings/{args.building_id}")
    set_last_request_id(ctx.state, envelope)
    return CommandResult("buildings show", extract_data(envelope), envelope.get("request_id"))


def cmd_buildings_shop(args: argparse.Namespace, ctx: RuntimeContext, client: APIClient) -> CommandResult:
    envelope = authenticated_request(ctx, client, "buildings shop", "GET", f"/buildings/{args.building_id}/shop-inventory")
    save_state(ctx.state_file, ctx.state)
    return CommandResult("buildings shop", extract_data(envelope), envelope.get("request_id"))


def cmd_buildings_purchase(args: argparse.Namespace, ctx: RuntimeContext, client: APIClient) -> CommandResult:
    envelope = authenticated_request(
        ctx,
        client,
        "buildings purchase",
        "POST",
        f"/buildings/{args.building_id}/purchase",
        body={"catalog_id": args.catalog_id},
    )
    data = extract_data(envelope)
    sync_action_payload(ctx.state, data if isinstance(data, dict) else None)
    save_state(ctx.state_file, ctx.state)
    return CommandResult("buildings purchase", data, envelope.get("request_id"))


def cmd_buildings_sell(args: argparse.Namespace, ctx: RuntimeContext, client: APIClient) -> CommandResult:
    envelope = authenticated_request(
        ctx,
        client,
        "buildings sell",
        "POST",
        f"/buildings/{args.building_id}/sell",
        body={"item_id": args.item_id},
    )
    data = extract_data(envelope)
    sync_action_payload(ctx.state, data if isinstance(data, dict) else None)
    save_state(ctx.state_file, ctx.state)
    return CommandResult("buildings sell", data, envelope.get("request_id"))


def cmd_buildings_salvage(args: argparse.Namespace, ctx: RuntimeContext, client: APIClient) -> CommandResult:
    envelope = authenticated_request(
        ctx,
        client,
        "buildings salvage",
        "POST",
        f"/buildings/{args.building_id}/salvage",
        body={"item_id": args.item_id},
    )
    data = extract_data(envelope)
    sync_action_payload(ctx.state, data if isinstance(data, dict) else None)
    save_state(ctx.state_file, ctx.state)
    return CommandResult("buildings salvage", data, envelope.get("request_id"))


def cmd_buildings_enhance(args: argparse.Namespace, ctx: RuntimeContext, client: APIClient) -> CommandResult:
    payload: dict[str, object] = {}
    if args.item_id:
        payload["item_id"] = args.item_id
    if args.slot:
        payload["slot"] = args.slot
    envelope = authenticated_request(ctx, client, "buildings enhance", "POST", f"/buildings/{args.building_id}/enhance", payload)
    data = extract_data(envelope)
    sync_action_payload(ctx.state, data if isinstance(data, dict) else None)
    save_state(ctx.state_file, ctx.state)
    return CommandResult("buildings enhance", data, envelope.get("request_id"))


def cmd_dungeons_list(args: argparse.Namespace, ctx: RuntimeContext, client: APIClient) -> CommandResult:
    envelope = client.request("dungeons list", "GET", "/dungeons")
    set_last_request_id(ctx.state, envelope)
    return CommandResult("dungeons list", extract_data(envelope), envelope.get("request_id"))


def cmd_dungeons_show(args: argparse.Namespace, ctx: RuntimeContext, client: APIClient) -> CommandResult:
    envelope = client.request("dungeons show", "GET", f"/dungeons/{args.dungeon_id}")
    set_last_request_id(ctx.state, envelope)
    return CommandResult("dungeons show", extract_data(envelope), envelope.get("request_id"))


def cmd_dungeons_enter(args: argparse.Namespace, ctx: RuntimeContext, client: APIClient) -> CommandResult:
    body: dict[str, Any] = {}
    if args.difficulty:
        body["difficulty"] = args.difficulty
    if args.potion_id:
        body["potion_loadout"] = list(args.potion_id)
    envelope = authenticated_request(ctx, client, "dungeons enter", "POST", f"/dungeons/{args.dungeon_id}/enter", body=body or None)
    data = extract_data(envelope)
    sync_run_view(ctx.state, data if isinstance(data, dict) else None)
    save_state(ctx.state_file, ctx.state)
    return CommandResult("dungeons enter", data, envelope.get("request_id"))


def cmd_dungeons_history(args: argparse.Namespace, ctx: RuntimeContext, client: APIClient) -> CommandResult:
    query = {
        "dungeon_id": args.dungeon_id,
        "difficulty": args.difficulty,
        "result": args.result,
        "limit": args.limit,
        "cursor": args.cursor,
    }
    envelope = authenticated_request(ctx, client, "dungeons history", "GET", "/me/runs", query=query)
    data = extract_data(envelope)
    save_state(ctx.state_file, ctx.state)
    return CommandResult("dungeons history", data, envelope.get("request_id"))


def cmd_dungeons_active(args: argparse.Namespace, ctx: RuntimeContext, client: APIClient) -> CommandResult:
    query = {"detail_level": args.detail_level} if args.detail_level else None
    envelope = authenticated_request(ctx, client, "dungeons active", "GET", "/me/runs/active", query=query)
    data = extract_data(envelope)
    sync_run_view(ctx.state, data if isinstance(data, dict) else None)
    save_state(ctx.state_file, ctx.state)
    return CommandResult("dungeons active", data, envelope.get("request_id"))


def cmd_dungeons_run(args: argparse.Namespace, ctx: RuntimeContext, client: APIClient) -> CommandResult:
    query = {"detail_level": args.detail_level} if args.detail_level else None
    envelope = authenticated_request(ctx, client, "dungeons run", "GET", f"/me/runs/{args.run_id}", query=query)
    data = extract_data(envelope)
    sync_run_view(ctx.state, data if isinstance(data, dict) else None)
    save_state(ctx.state_file, ctx.state)
    return CommandResult("dungeons run", data, envelope.get("request_id"))


def cmd_dungeons_claim(args: argparse.Namespace, ctx: RuntimeContext, client: APIClient) -> CommandResult:
    envelope = authenticated_request(ctx, client, "dungeons claim", "POST", f"/me/runs/{args.run_id}/claim")
    data = extract_data(envelope)
    sync_run_view(ctx.state, data if isinstance(data, dict) else None)
    pending = [item for item in ctx.state.get("pending_claim_run_ids", []) if item != args.run_id]
    ctx.state["pending_claim_run_ids"] = pending
    save_state(ctx.state_file, ctx.state)
    return CommandResult("dungeons claim", data, envelope.get("request_id"))


def cmd_arena_signup(args: argparse.Namespace, ctx: RuntimeContext, client: APIClient) -> CommandResult:
    envelope = authenticated_request(ctx, client, "arena signup", "POST", "/arena/signup")
    data = extract_data(envelope)
    sync_action_payload(ctx.state, data if isinstance(data, dict) else None)
    save_state(ctx.state_file, ctx.state)
    return CommandResult("arena signup", data, envelope.get("request_id"))


def cmd_arena_current(args: argparse.Namespace, ctx: RuntimeContext, client: APIClient) -> CommandResult:
    envelope = authenticated_request(ctx, client, "arena current", "GET", "/arena/current")
    save_state(ctx.state_file, ctx.state)
    return CommandResult("arena current", extract_data(envelope), envelope.get("request_id"))


def cmd_arena_entries(args: argparse.Namespace, ctx: RuntimeContext, client: APIClient) -> CommandResult:
    query: dict[str, Any] = {}
    if args.limit is not None:
        query["limit"] = args.limit
    if args.cursor:
        query["cursor"] = args.cursor
    envelope = authenticated_request(ctx, client, "arena entries", "GET", "/arena/entries", query=query or None)
    save_state(ctx.state_file, ctx.state)
    return CommandResult("arena entries", extract_data(envelope), envelope.get("request_id"))


def cmd_arena_leaderboard(args: argparse.Namespace, ctx: RuntimeContext, client: APIClient) -> CommandResult:
    envelope = authenticated_request(ctx, client, "arena leaderboard", "GET", "/arena/leaderboard")
    save_state(ctx.state_file, ctx.state)
    return CommandResult("arena leaderboard", extract_data(envelope), envelope.get("request_id"))


def cmd_action(args: argparse.Namespace, ctx: RuntimeContext, client: APIClient) -> CommandResult:
    action_args = parse_key_value_pairs(args.action_arg or [])
    payload: dict[str, Any] = {
        "action_type": args.action_type,
        "action_args": action_args,
    }
    if args.client_turn_id:
        payload["client_turn_id"] = args.client_turn_id
    envelope = authenticated_request(ctx, client, "action", "POST", "/me/actions", body=payload)
    data = extract_data(envelope)
    sync_action_payload(ctx.state, data if isinstance(data, dict) else None)
    save_state(ctx.state_file, ctx.state)
    return CommandResult("action", data, envelope.get("request_id"))


def cmd_raw(args: argparse.Namespace, ctx: RuntimeContext, client: APIClient) -> CommandResult:
    body = json.loads(args.body_json) if args.body_json else None
    query = parse_key_value_pairs(args.query or []) if args.query else None
    token = ensure_access_token(ctx, client) if args.auth else None
    envelope = client.request(
        "raw",
        args.method.upper(),
        args.path,
        body=body,
        query=query,
        token=token,
        base=args.base,
    )
    set_last_request_id(ctx.state, envelope)
    save_state(ctx.state_file, ctx.state)
    return CommandResult("raw", envelope, envelope.get("request_id"))


def build_parser() -> argparse.ArgumentParser:
    parser = argparse.ArgumentParser(prog="clawgame", description="Bundled ClawGame CLI for OpenClaw and other agents")
    parser.add_argument("--api-base", default=DEFAULT_API_BASE)
    parser.add_argument("--observer-origin", default=DEFAULT_OBSERVER_ORIGIN)
    parser.add_argument("--state-file", default=DEFAULT_STATE_FILE)
    parser.add_argument("--access-token")
    parser.add_argument("--timeout-seconds", type=float, default=15.0)
    parser.add_argument("--pretty", action="store_true")

    subparsers = parser.add_subparsers(dest="command", required=True)

    bootstrap = subparsers.add_parser("bootstrap", help="establish or resume a session and optionally ensure a character")
    bootstrap.add_argument("--bot-name")
    bootstrap.add_argument("--password")
    bootstrap.add_argument("--character-name")
    bootstrap.add_argument("--gender", choices=("male", "female"))
    bootstrap.add_argument("--class", dest="character_class")
    bootstrap.add_argument("--weapon-style")
    bootstrap.add_argument("--register-if-needed", choices=("true", "false"))
    bootstrap.set_defaults(func=cmd_bootstrap)

    register = subparsers.add_parser("register", help="create an account with challenge-based auth")
    register.add_argument("--bot-name")
    register.add_argument("--password")
    register.set_defaults(func=cmd_register)

    login = subparsers.add_parser("login", help="login with challenge-based auth")
    login.add_argument("--bot-name")
    login.add_argument("--password")
    login.set_defaults(func=cmd_login)

    refresh = subparsers.add_parser("refresh", help="refresh the current session")
    refresh.set_defaults(func=cmd_refresh)

    me = subparsers.add_parser("me", help="fetch account and character summary")
    me.set_defaults(func=cmd_me)

    planner = subparsers.add_parser("planner", help="fetch planner summary")
    planner.add_argument("--region-id")
    planner.set_defaults(func=cmd_planner)

    state = subparsers.add_parser("state", help="fetch full current state")
    state.set_defaults(func=cmd_state)

    actions = subparsers.add_parser("actions", help="list valid current actions")
    actions.set_defaults(func=cmd_actions)

    field = subparsers.add_parser("field", help="resolve a field interaction in the current region")
    field_sub = field.add_subparsers(dest="field_command", required=True)
    field_hunt = field_sub.add_parser("hunt", help="resolve a hunt encounter in the current field region")
    field_hunt.set_defaults(func=cmd_field, approach="hunt")
    field_gather = field_sub.add_parser("gather", help="resolve a gathering encounter in the current field region")
    field_gather.set_defaults(func=cmd_field, approach="gather")
    field_curio = field_sub.add_parser("curio", help="resolve a curio encounter in the current field region")
    field_curio.set_defaults(func=cmd_field, approach="curio")

    regions = subparsers.add_parser("regions", help="world region discovery")
    regions_sub = regions.add_subparsers(dest="regions_command", required=True)
    regions_list = regions_sub.add_parser("list", help="list regions")
    regions_list.set_defaults(func=cmd_regions_list)
    regions_show = regions_sub.add_parser("show", help="show a region")
    regions_show.add_argument("--region-id", required=True)
    regions_show.set_defaults(func=cmd_regions_show)

    travel = subparsers.add_parser("travel", help="travel to another region")
    travel.add_argument("--region-id", required=True)
    travel.set_defaults(func=cmd_travel)

    quests = subparsers.add_parser("quests", help="quest operations")
    quests_sub = quests.add_subparsers(dest="quests_command", required=True)
    quests_list = quests_sub.add_parser("list", help="list quests")
    quests_list.set_defaults(func=cmd_quests_list)
    quests_show = quests_sub.add_parser("show", help="show quest runtime detail")
    quests_show.add_argument("--quest-id", required=True)
    quests_show.set_defaults(func=cmd_quests_show)
    quests_accept = quests_sub.add_parser("accept", help="accept a quest")
    quests_accept.add_argument("--quest-id", required=True)
    quests_accept.set_defaults(func=cmd_quests_accept)
    quests_choice = quests_sub.add_parser("choice", help="submit a quest choice")
    quests_choice.add_argument("--quest-id", required=True)
    quests_choice.add_argument("--choice-key", required=True)
    quests_choice.set_defaults(func=cmd_quests_choice)
    quests_interact = quests_sub.add_parser("interact", help="submit a quest interaction")
    quests_interact.add_argument("--quest-id", required=True)
    quests_interact.add_argument("--interaction", required=True)
    quests_interact.set_defaults(func=cmd_quests_interact)
    quests_submit = quests_sub.add_parser("submit", help="submit a quest")
    quests_submit.add_argument("--quest-id", required=True)
    quests_submit.set_defaults(func=cmd_quests_submit)
    quests_reroll = quests_sub.add_parser("reroll", help="reroll the quest board")
    quests_reroll.add_argument("--confirm-cost", choices=("true", "false"), default="true")
    quests_reroll.set_defaults(func=cmd_quests_reroll)

    inventory = subparsers.add_parser("inventory", help="inspect inventory and equipment")
    inventory.set_defaults(func=cmd_inventory)

    equipment = subparsers.add_parser("equipment", help="equipment actions")
    equipment_sub = equipment.add_subparsers(dest="equipment_command", required=True)
    equip = equipment_sub.add_parser("equip", help="equip an item")
    equip.add_argument("--item-id", required=True)
    equip.set_defaults(func=cmd_equipment_equip)
    unequip = equipment_sub.add_parser("unequip", help="unequip a slot")
    unequip.add_argument("--slot", required=True)
    unequip.set_defaults(func=cmd_equipment_unequip)

    buildings = subparsers.add_parser("buildings", help="building actions")
    buildings_sub = buildings.add_subparsers(dest="buildings_command", required=True)
    buildings_show = buildings_sub.add_parser("show", help="show a building")
    buildings_show.add_argument("--building-id", required=True)
    buildings_show.set_defaults(func=cmd_buildings_show)
    buildings_shop = buildings_sub.add_parser("shop", help="show shop inventory")
    buildings_shop.add_argument("--building-id", required=True)
    buildings_shop.set_defaults(func=cmd_buildings_shop)
    buildings_purchase = buildings_sub.add_parser("purchase", help="purchase an item")
    buildings_purchase.add_argument("--building-id", required=True)
    buildings_purchase.add_argument("--catalog-id", required=True)
    buildings_purchase.set_defaults(func=cmd_buildings_purchase)
    buildings_sell = buildings_sub.add_parser("sell", help="sell an item")
    buildings_sell.add_argument("--building-id", required=True)
    buildings_sell.add_argument("--item-id", required=True)
    buildings_sell.set_defaults(func=cmd_buildings_sell)
    buildings_salvage = buildings_sub.add_parser("salvage", help="salvage an item")
    buildings_salvage.add_argument("--building-id", required=True)
    buildings_salvage.add_argument("--item-id", required=True)
    buildings_salvage.set_defaults(func=cmd_buildings_salvage)
    buildings_enhance = buildings_sub.add_parser("enhance", help="enhance at a building")
    buildings_enhance.add_argument("--building-id", required=True)
    buildings_enhance.add_argument("--slot")
    buildings_enhance.add_argument("--item-id")
    buildings_enhance.set_defaults(func=cmd_buildings_enhance)

    dungeons = subparsers.add_parser("dungeons", help="dungeon operations")
    dungeons_sub = dungeons.add_subparsers(dest="dungeons_command", required=True)
    dungeons_list = dungeons_sub.add_parser("list", help="list dungeons")
    dungeons_list.set_defaults(func=cmd_dungeons_list)
    dungeons_show = dungeons_sub.add_parser("show", help="show a dungeon")
    dungeons_show.add_argument("--dungeon-id", required=True)
    dungeons_show.set_defaults(func=cmd_dungeons_show)
    dungeons_enter = dungeons_sub.add_parser("enter", help="enter a dungeon")
    dungeons_enter.add_argument("--dungeon-id", required=True)
    dungeons_enter.add_argument("--difficulty", choices=("easy", "hard", "nightmare"))
    dungeons_enter.add_argument("--potion-id", action="append", default=[])
    dungeons_enter.set_defaults(func=cmd_dungeons_enter)
    dungeons_history = dungeons_sub.add_parser("history", help="list historical dungeon runs")
    dungeons_history.add_argument("--dungeon-id")
    dungeons_history.add_argument("--difficulty", choices=("easy", "hard", "nightmare"))
    dungeons_history.add_argument("--result", choices=("cleared", "failed", "abandoned", "expired"))
    dungeons_history.add_argument("--limit", type=int)
    dungeons_history.add_argument("--cursor")
    dungeons_history.set_defaults(func=cmd_dungeons_history)
    dungeons_active = dungeons_sub.add_parser("active", help="show active run")
    dungeons_active.add_argument("--detail-level", choices=("compact", "standard", "verbose"))
    dungeons_active.set_defaults(func=cmd_dungeons_active)
    dungeons_run = dungeons_sub.add_parser("run", help="show a run")
    dungeons_run.add_argument("--run-id", required=True)
    dungeons_run.add_argument("--detail-level", choices=("compact", "standard", "verbose"))
    dungeons_run.set_defaults(func=cmd_dungeons_run)
    dungeons_claim = dungeons_sub.add_parser("claim", help="claim dungeon rewards")
    dungeons_claim.add_argument("--run-id", required=True)
    dungeons_claim.set_defaults(func=cmd_dungeons_claim)

    arena = subparsers.add_parser("arena", help="arena operations")
    arena_sub = arena.add_subparsers(dest="arena_command", required=True)
    arena_signup = arena_sub.add_parser("signup", help="sign up for arena")
    arena_signup.set_defaults(func=cmd_arena_signup)
    arena_current = arena_sub.add_parser("current", help="show current arena state")
    arena_current.set_defaults(func=cmd_arena_current)
    arena_entries = arena_sub.add_parser("entries", help="list arena entrants with pagination")
    arena_entries.add_argument("--limit", type=int)
    arena_entries.add_argument("--cursor")
    arena_entries.set_defaults(func=cmd_arena_entries)
    arena_leaderboard = arena_sub.add_parser("leaderboard", help="show arena leaderboard")
    arena_leaderboard.set_defaults(func=cmd_arena_leaderboard)

    action = subparsers.add_parser("action", help="fallback wrapper for POST /me/actions")
    action.add_argument("--action-type", required=True)
    action.add_argument("--action-arg", action="append", default=[])
    action.add_argument("--client-turn-id")
    action.set_defaults(func=cmd_action)

    raw = subparsers.add_parser("raw", help="generic HTTP fallback")
    raw.add_argument("--method", required=True)
    raw.add_argument("--path", required=True)
    raw.add_argument("--query", action="append", default=[])
    raw.add_argument("--body-json")
    raw.add_argument("--auth", action="store_true")
    raw.add_argument("--base", choices=("api", "observer"), default="api")
    raw.set_defaults(func=cmd_raw)

    return parser


def main(argv: list[str] | None = None) -> int:
    parser = build_parser()
    args = parser.parse_args(argv)
    state_file = Path(args.state_file)
    state = load_state(state_file)
    ctx = RuntimeContext(
        api_base=args.api_base,
        observer_origin=args.observer_origin,
        state_file=state_file,
        pretty=args.pretty,
        timeout_seconds=args.timeout_seconds,
        explicit_access_token=args.access_token,
        state=state,
    )
    client = APIClient(args.api_base, args.observer_origin, args.timeout_seconds)

    try:
        handler: Callable[[argparse.Namespace, RuntimeContext, APIClient], CommandResult] = args.func
        result = handler(args, ctx, client)
        emit_json(success_payload(ctx, result), pretty=ctx.pretty)
        return 0
    except CommandFailure as exc:
        emit_json(exc.payload, pretty=ctx.pretty)
        return exc.exit_code


if __name__ == "__main__":
    sys.exit(main())
