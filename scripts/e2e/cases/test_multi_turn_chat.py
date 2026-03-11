"""Test: Multi-turn chat with real Claude responses.

Verifies:
1. User sends message and sees it displayed
2. Claude actually responds (assistant message appears)
3. Second message works (multi-turn in same session)
4. After page refresh, all messages are restored from localStorage
"""

import logging
from pathlib import Path
from playwright.sync_api import Page, expect


def run(page: Page, base_url: str, out_dir: Path, logger: logging.Logger) -> bool:
    logger.info("Step 1: Navigate and wait for connection")
    page.goto(base_url)
    expect(page.locator("text=已连接")).to_be_visible(timeout=10000)
    page.screenshot(path=str(out_dir / "screenshots" / "01_connected.png"))

    logger.info("Step 2: Send first message")
    textarea = page.locator("textarea")
    textarea.fill("Reply with only the word 'pong'. Nothing else.")
    page.screenshot(path=str(out_dir / "screenshots" / "02_message_typed.png"))

    send_btn = page.locator("button[title='发送']")
    send_btn.click()
    page.wait_for_timeout(500)
    page.screenshot(path=str(out_dir / "screenshots" / "03_message_sent.png"))

    logger.info("Step 3: Verify user message appears")
    user_msg = page.locator("text=Reply with only the word")
    expect(user_msg).to_be_visible(timeout=5000)
    page.screenshot(path=str(out_dir / "screenshots" / "04_user_msg_visible.png"))

    logger.info("Step 4: Wait for Claude's response (must receive a reply)")
    page.locator("textarea:not([disabled])").wait_for(timeout=60000)
    page.wait_for_timeout(500)
    page.screenshot(path=str(out_dir / "screenshots" / "05_response_received.png"))

    logger.info("Step 5: Verify assistant message exists")
    assistant_blocks = page.locator("div:has(> div.text-xs:has-text('Claude'))")
    count = assistant_blocks.count()
    logger.info(f"Found {count} assistant message blocks")
    assert count >= 1, f"Expected at least 1 assistant message, found {count}. Claude did not respond!"
    page.screenshot(path=str(out_dir / "screenshots" / "06_assistant_visible.png"))

    logger.info("Step 6: Send second message (multi-turn)")
    textarea = page.locator("textarea")
    textarea.fill("Reply with only the word 'ping'. Nothing else.")
    send_btn = page.locator("button[title='发送']")
    send_btn.click()
    page.wait_for_timeout(500)
    page.screenshot(path=str(out_dir / "screenshots" / "07_second_msg_sent.png"))

    logger.info("Step 7: Verify second user message appears")
    user_msg2 = page.locator("text=Reply with only the word 'ping'")
    expect(user_msg2).to_be_visible(timeout=5000)

    logger.info("Step 8: Wait for second Claude response")
    page.locator("textarea:not([disabled])").wait_for(timeout=60000)
    page.wait_for_timeout(500)
    page.screenshot(path=str(out_dir / "screenshots" / "08_second_response.png"))

    logger.info("Step 9: Verify multiple assistant messages now")
    assistant_blocks2 = page.locator("div:has(> div.text-xs:has-text('Claude'))")
    count2 = assistant_blocks2.count()
    logger.info(f"Found {count2} assistant message blocks after second turn")
    assert count2 >= 2, f"Expected at least 2 assistant messages after multi-turn, found {count2}"
    page.screenshot(path=str(out_dir / "screenshots" / "09_multi_turn_done.png"))

    logger.info("Step 10: Verify localStorage has session data")
    session_data = page.evaluate("localStorage.getItem('claude-webui-session')")
    assert session_data is not None, "localStorage should have session data"
    import json
    parsed = json.loads(session_data)
    assert parsed.get("sessionId"), "sessionId should be set in localStorage"
    msg_count = len(parsed.get("messages", []))
    logger.info(f"localStorage has {msg_count} messages, sessionId={parsed['sessionId'][:12]}...")
    assert msg_count >= 4, f"Expected at least 4 messages (2 user + 2 assistant), found {msg_count}"
    page.screenshot(path=str(out_dir / "screenshots" / "10_session_stored.png"))

    logger.info("Step 11: Refresh page and verify messages restored")
    page.reload()
    expect(page.locator("text=已连接")).to_be_visible(timeout=10000)
    page.wait_for_timeout(500)

    expect(page.locator("text=Reply with only the word 'pong'").first).to_be_visible(timeout=5000)
    expect(page.locator("text=Reply with only the word 'ping'").first).to_be_visible(timeout=5000)
    page.screenshot(path=str(out_dir / "screenshots" / "11_restored_after_refresh.png"))

    logger.info("All checks passed")
    return True
