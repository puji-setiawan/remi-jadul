# 0012 — Pseudo-3D tabletop presentation for MVP

## Status

Accepted

## Context

Remi Jadul needs a modern tabletop feel with depth and physical card movement, but a full 3D engine would add unnecessary runtime, asset, and interaction complexity for a 2D card game.

## Decision

The Flutter client uses pseudo-3D presentation for the MVP. Perspective transforms, rotation, scale, elevation, shadows, and card movement may create the tabletop depth effect. Unity, Godot, free camera controls, 3D meshes, and complex physics are out of MVP scope.

## Consequences

The FE implementation owns only presentation and animation. The server remains authoritative for card identity, ownership, combinations, turns, and scores. The design must preserve readability and provide a reduced-motion path that minimizes or disables perspective movement.
