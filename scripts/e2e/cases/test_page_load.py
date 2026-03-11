"""Test: Page loads correctly and WebSocket connects."""

import logging
from pathlib import Path
from playwright.sync_api import Page, expect


def run(page: Page, base_url: str, out_dir: Path, logger: logging.Logger) -> bool:
    logger.info("Step 1: Navigate to the app")
    page.goto(base_url)
    page.screenshot(path=str(out_dir / "screenshots" / "01_navigate.png"))

    logger.info("Step 2: Verify page title / heading")
    expect(page.get_by_role("heading", name="Claude Code WebUI")).to_be_visible(timeout=5000)
    page.screenshot(path=str(out_dir / "screenshots" / "02_page_loaded.png"))

    logger.info("Step 3: Wait for WebSocket connection")
    connected = page.locator("text=已连接")
    expect(connected).to_be_visible(timeout=10000)
    page.screenshot(path=str(out_dir / "screenshots" / "03_ws_connected.png"))

    logger.info("Step 4: Verify empty state / welcome message")
    welcome = page.locator("text=发送消息开始对话")
    expect(welcome).to_be_visible(timeout=5000)
    page.screenshot(path=str(out_dir / "screenshots" / "04_welcome.png"))

    logger.info("Step 5: Verify input area is present")
    textarea = page.locator("textarea")
    expect(textarea).to_be_visible()
    expect(textarea).to_be_enabled()
    page.screenshot(path=str(out_dir / "screenshots" / "05_input_ready.png"))

    logger.info("All checks passed")
    return True
