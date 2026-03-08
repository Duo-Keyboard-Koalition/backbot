"""Chat channels module with plugin architecture."""

from darci.channels.base import BaseChannel
from darci.channels.manager import ChannelManager

__all__ = ["BaseChannel", "ChannelManager"]
