"""Test: Session persistence across page refresh with real responses.

Verifies that:
1. Send messages and receive actual Claude responses
2. After page refresh, all messages (user + assistant) are restored
3. Can continue chatting after restore
"""

import json
import logging
from pathlib import Path
from playwright.sync_api import Page, expect


def run(page: Page, base_url: str, out_dir: Path, logger: logging.Logger) -> bool:
    logger.info("Step 1: Navigate and connect")
    page.goto(base_url)
    expect(page.locator("text=已连接")).to_be_visible(timeout=10000)
    page.screenshot(path=str(out_dir / "screenshots" / "01_connected.png"))

    logger.info("Step 2: Send first message")
    textarea = page.locator("textarea")
    textarea.fill("Respond with only: alpha-ok")
    page.locator("button[title='发送']").click()
    page.wait_for_timeout(500)
    page.screenshot(path=str(out_dir / "screenshots" / "02_msg_sent.png"))

    logger.info("Step 3: Verify user message displayed")
    expect(page.locator("text=Respond with only: alpha-ok")).to_be_visible(timeout=5000)

    logger.info("Step 4: Wait for Claude response")
    page.locator("textarea:not([disabled])").wait_for(timeout=60000)
    page.wait_for_timeout(500)
    page.screenshot(path=str(out_dir / "screenshots" / "03_first_response.png"))

    logger.info("Step 5: Verify assistant actually responded")
    assistant_blocks = page.locator("div:has(> div.text-xs:has-text('Claude'))")
    count = assistant_blocks.count()
    assert count >= 1, f"Claude did not respond! Found {count} assistant blocks"
    logger.info(f"First response received, {count} assistant blocks")

    logger.info("Step 6: Send second message")
    textarea = page.locator("textarea")
    textarea.fill("Respond with only: beta-ok")
    page.locator("button[title='发送']").click()
    page.wait_for_timeout(500)
    page.screenshot(path=str(out_dir / "screenshots" / "04_second_sent.png"))

    logger.info("Step 7: Wait for second Claude response")
    page.locator("textarea:not([disabled])").wait_for(timeout=60000)
    page.wait_for_timeout(500)
    page.screenshot(path=str(out_dir / "screenshots" / "05_second_response.png"))

    assistant_blocks2 = page.locator("div:has(> div.text-xs:has-text('Claude'))")
    count2 = assistant_blocks2.count()
    assert count2 >= 2, f"Second response not received! Found {count2} assistant blocks"
    logger.info(f"Second response received, {count2} assistant blocks")

    logger.info("Step 8: Verify localStorage")
    data = page.evaluate("localStorage.getItem('claude-webui-session')")
    assert data is not None, "No session data in localStorage"
    parsed = json.loads(data)
    msg_count = len(parsed.get("messages", []))
    session_id = parsed.get("sessionId", "")
    logger.info(f"localStorage: {msg_count} messages, sessionId={session_id[:12]}...")
    assert msg_count >= 4, f"Expected >= 4 messages, found {msg_count}"
    assert session_id, "sessionId should not be empty"
    page.screenshot(path=str(out_dir / "screenshots" / "06_localstorage.png"))

    logger.info("Step 9: Refresh the page")
    page.reload()
    page.wait_for_timeout(1000)
    page.screenshot(path=str(out_dir / "screenshots" / "07_after_refresh.png"))

    logger.info("Step 10: Wait for reconnection")
    expect(page.locator("text=已连接")).to_be_visible(timeout=15000)

    logger.info("Step 11: Verify all messages restored after refresh")
    expect(page.locator("text=Respond with only: alpha-ok")).to_be_visible(timeout=5000)
    expect(page.locator("text=Respond with only: beta-ok")).to_be_visible(timeout=5000)

    restored_assistants = page.locator("div:has(> div.text-xs:has-text('Claude'))")
    restored_count = restored_assistants.count()
    assert restored_count >= 2, f"Assistant messages not restored! Found {restored_count}"
    page.screenshot(path=str(out_dir / "screenshots" / "08_restored.png"))

    logger.info("Step 12: Verify can still type after restore")
    textarea = page.locator("textarea")
    expect(textarea).to_be_enabled(timeout=10000)
    page.screenshot(path=str(out_dir / "screenshots" / "09_can_type.png"))

    logger.info("All checks passed")
    return True
