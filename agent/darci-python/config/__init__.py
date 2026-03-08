"""Configuration module for darci."""

from darci.config.loader import load_config, get_config_path
from darci.config.schema import Config

__all__ = ["Config", "load_config", "get_config_path"]
