# Visibility Matrix v1

| Data | Owner | Other players | Public state |
|---|---:|---:|---:|
| Own hand card IDs/ranks/suits | full | no | no |
| Opponent hand | no | own only | count only |
| Complete discard pile | full | full | yes |
| Active joker rank | yes | yes | yes |
| Physical joker marker | no | no | no |
| Closed-deck order | no | no | no |
| Locked table combinations | full | full | yes |
| Current turn/phase | yes | yes | yes |
| Completed-round cumulative scores | yes | yes | yes |
| Current uncommitted hand score | private if needed | no | no |

Room updates are personalized projections. A private room code is never included in a broadcast payload; it is returned only to the creating host or the player who joined with that code.

Contract tests must assert both allowed fields and forbidden-field absence.
