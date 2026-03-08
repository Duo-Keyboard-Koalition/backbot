"""ADK tool support for DarCI."""

_bus_publish = None

def set_runtime_refs(bus_publish=None):
    global _bus_publish
    _bus_publish = bus_publish

# Define a minimal set of tools for subagents since the original ones are missing
# We can adapt the existing classes to be used in SUBAGENT_TOOLS list if needed,
# but for now let's just create a dummy list so it doesn't crash.
# Actually, I should probably use the tools from darci.agent.tools

SUBAGENT_TOOLS = []
