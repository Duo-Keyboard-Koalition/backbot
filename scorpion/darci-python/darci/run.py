"""DarCI agent entry point — Darci configured as a DARCI project manager."""
import asyncio
import shutil
from pathlib import Path

from darci.agent.loop import AdkAgentLoop
from darci.bus.queue import MessageBus
from darci.config.loader import load_config
from darci.providers.gemini_provider import GeminiProvider
from darci.config.darci import DarciConfig
from darci.state.store import TaskStore
from darci.agent.tools.darci.register import register_darci_tools

DARCI_WORKSPACE = Path("~/.darci/workspace").expanduser()
_PKG_DIR = Path(__file__).parent
SOUL_SRC = _PKG_DIR / "workspace" / "SOUL.md"
SKILL_SRC = _PKG_DIR / "workspace" / "skills" / "darci" / "SKILL.md"


def _init_workspace(ws: Path) -> None:
    """Copy DarCI identity files to the Darci workspace on first run."""
    ws.mkdir(parents=True, exist_ok=True)

    soul_dst = ws / "SOUL.md"
    if not soul_dst.exists() and SOUL_SRC.exists():
        shutil.copy(SOUL_SRC, soul_dst)

    skill_dst = ws / "skills" / "darci" / "SKILL.md"
    skill_dst.parent.mkdir(parents=True, exist_ok=True)
    if not skill_dst.exists() and SKILL_SRC.exists():
        shutil.copy(SKILL_SRC, skill_dst)


async def main() -> None:
    config_darci = load_config()
    darci_config = DarciConfig()
    store = TaskStore(darci_config)

    _init_workspace(DARCI_WORKSPACE)

    bus = MessageBus()
    provider = GeminiProvider(
        api_key=config_darci.providers.gemini.api_key,
        default_model=config_darci.agents.defaults.model,
    )

    agent = AdkAgentLoop(
        bus=bus,
        provider=provider,
        workspace=DARCI_WORKSPACE,
        model=config_darci.agents.defaults.model,
        temperature=config_darci.agents.defaults.temperature,
        max_tokens=config_darci.agents.defaults.max_tokens,
        max_iterations=config_darci.agents.defaults.max_tool_iterations,
        memory_window=config_darci.agents.defaults.memory_window,
    )

    register_darci_tools(agent, darci_config, store)

    print("DarCI ready — DARCI project manager for AI agents")
    print(f"State: {darci_config.state_dir}")
    print(f"Bridge: {darci_config.bridge_local_url}")
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
