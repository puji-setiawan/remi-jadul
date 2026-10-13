# Remi Jadul — Visual Direction

This file is the design direction for Flutter/UI work. It is not a substitute for product requirements or game rules.

## Product feeling

Remi Jadul should feel like a familiar tabletop card game at home: warm, tactile, readable, calm, and slightly nostalgic. It should not look like a generic SaaS dashboard, casino advertisement, or crypto game.

## Visual language

- Use a deep felt-green table as the primary play surface.
- Use warm ivory/card-paper surfaces and restrained red/black suit colors.
- Prefer solid colors, subtle texture, and depth from spacing/shadow over gradients, glassmorphism, or excessive glow.
- Use clear card geometry and readable rank/suit marks at phone size.
- Keep the active player, valid action, selected card, and error state visually unmistakable.
- Use motion only to explain a card movement, turn change, or state transition; never as decoration that delays play.

## Layout and interaction

- Design for one-handed Android use and narrow screens first.
- Keep the current turn and primary action visible without hunting through menus.
- Never hide critical game state behind decorative panels.
- Preserve card proportions and tap targets when the hand wraps or scrolls.
- Give loading, empty, error, reconnecting, disabled, and success states explicit designs.

## Accessibility

- Do not use color alone to communicate suit, turn, validity, or error.
- Maintain readable contrast and scalable text.
- Ensure tap targets are practical on a phone.
- Provide semantic labels for cards and actions.
- Respect reduced-motion preferences where the platform supports them.

## Anti-slop constraints

- No invented statistics, fake activity counters, arbitrary badges, or decorative dashboard metrics.
- No generic “Unlock your potential”, “Level up”, or similar filler copy.
- No gradient hero banner, sparkle logo, or rounded-card dashboard pattern unless a concrete product need is documented.
- Use real game state and real labels. If data does not exist, show an honest empty state.
