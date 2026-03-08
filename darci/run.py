"""DarCI agent entry point — Scorpion configured as a DARCI project manager."""
import asyncio
import shutil
from pathlib import Path

from scorpion.agent.loop import AdkAgentLoop
from scorpion.bus.queue import MessageBus
from scorpion.config.loader import load_config
from scorpion.providers.gemini_provider import GeminiProvider

from darci.config import DarciConfig
from darci.state.store import TaskStore
from darci.tools.register import register_darci_tools

DARCI_WORKSPACE = Path("~/.scorpion/darci/workspace").expanduser()
_PKG_DIR = Path(__file__).parent
SOUL_SRC = _PKG_DIR / "workspace" / "SOUL.md"
SKILL_SRC = _PKG_DIR / "workspace" / "skills" / "darci" / "SKILL.md"


def _init_workspace(ws: Path) -> None:
    """Copy DarCI identity files to the Scorpion workspace on first run."""
    ws.mkdir(parents=True, exist_ok=True)

    soul_dst = ws / "SOUL.md"
    if not soul_dst.exists() and SOUL_SRC.exists():
        shutil.copy(SOUL_SRC, soul_dst)

    skill_dst = ws / "skills" / "darci" / "SKILL.md"
    skill_dst.parent.mkdir(parents=True, exist_ok=True)
    if not skill_dst.exists() and SKILL_SRC.exists():
        shutil.copy(SKILL_SRC, skill_dst)


async def main() -> None:
    config_sc = load_config()
    config_darci = DarciConfig()
    store = TaskStore(config_darci)

    _init_workspace(DARCI_WORKSPACE)

    bus = MessageBus()
    provider = GeminiProvider(
        api_key=config_sc.providers.gemini.api_key,
        default_model=config_sc.agents.defaults.model,
    )

    agent = AdkAgentLoop(
        bus=bus,
        provider=provider,
        workspace=DARCI_WORKSPACE,
        model=config_sc.agents.defaults.model,
        temperature=config_sc.agents.defaults.temperature,
        max_tokens=config_sc.agents.defaults.max_tokens,
        max_iterations=config_sc.agents.defaults.max_tool_iterations,
        memory_window=config_sc.agents.defaults.memory_window,
    )

    register_darci_tools(agent, config_darci, store)

    print("DarCI ready — DARCI project manager for AI agents")
    print(f"State: {config_darci.state_dir}")
    print(f"Bridge: {config_darci.bridge_local_url}")
    print("Type your command. Ctrl-C to exit.\n")

    while True:
        try:
            user_input = input("darci> ").strip()
        except (EOFError, KeyboardInterrupt):
            print("\nDarCI shutting down.")
            break
        if not user_input:
            continue
        response = await agent.process_direct(user_input)
        print(f"\n{response}\n")
