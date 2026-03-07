"""
scorpion - A lightweight AI agent framework
"""

__version__ = "0.1.4.post3"
__logo__ = "🐈"

# Ensure UTF-8 encoding for Windows console
import sys
import codecs
if sys.platform == "win32":
    sys.stdout = codecs.getwriter("utf-8")(sys.stdout.buffer, "strict")
    sys.stderr = codecs.getwriter("utf-8")(sys.stderr.buffer, "strict")
