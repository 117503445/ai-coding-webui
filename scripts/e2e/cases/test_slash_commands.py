"""Test: Slash command menu appears and commands work."""

import logging
from pathlib import Path
from playwright.sync_api import Page, expect


def run(page: Page, base_url: str, out_dir: Path, logger: logging.Logger) -> bool:
    logger.info("Step 1: Navigate and wait for connection")
    page.goto(base_url)
    expect(page.locator("text=已连接")).to_be_visible(timeout=10000)
    page.screenshot(path=str(out_dir / "screenshots" / "01_connected.png"))

    logger.info("Step 2: Type '/' to trigger command menu")
    textarea = page.locator("textarea")
    textarea.fill("/")
    page.wait_for_timeout(300)
    page.screenshot(path=str(out_dir / "screenshots" / "02_slash_typed.png"))

    logger.info("Step 3: Verify command menu appears")
    menu = page.locator("text=/new")
    expect(menu).to_be_visible(timeout=3000)
    page.screenshot(path=str(out_dir / "screenshots" / "03_command_menu.png"))

    logger.info("Step 4: Filter commands by typing '/n'")
    textarea.fill("/n")
    page.wait_for_timeout(300)
    new_cmd = page.locator("button:has-text('/new')")
    expect(new_cmd).to_be_visible(timeout=3000)
    page.screenshot(path=str(out_dir / "screenshots" / "04_filtered_menu.png"))

    logger.info("Step 5: Click /new command")
    new_cmd.click()
    page.wait_for_timeout(500)
    page.screenshot(path=str(out_dir / "screenshots" / "05_new_executed.png"))

    logger.info("Step 6: Verify textarea is cleared after command")
    expect(textarea).to_have_value("")
    page.screenshot(path=str(out_dir / "screenshots" / "06_cleared.png"))

    logger.info("All checks passed")
    return True
