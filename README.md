# tic-tac-idle
An idle game for playing tic-tac-toe and ramping up the number of games and learning more about the meaning of the name tic-tac-toe
and the friends we made along the way

## Instructions for Play
Should be none needed, head over [here](https://tictac.djacoby.dev/) and start clicking around and seeing what you find!

# My Approach
In my screen recording I speak for a bit on my approach and planning phase, so more detail can be found there.

However and in general I was hoping to do the following in priority order:
1. Create something that couldn't be done in 90 mins without AI assistance. I suppose to showcase the increased productivity factor
2. Implement a fun gimmick that either made the game interesting from an implementation perspective or simply made it surprisingly fun
3. Provide a fully polished and "productionalized" product, something truly complete-enough and with empathy for the user running the game, hence the browser implementation

# AI Tooling

Pretty much just relied on Claude Code here, sonnet 4.6 has served me well for most of my coding activities, especially the smaller projects.

For larger efforts I definitely find paying close attention to the skills used and MCP servers the agentry interacts with is very important. However,
For a time sensitive exercise like this, I just wanted to hit the ground running and respect the minimal pre-work request

# Planning mishaps

* I noted the following areas that I would improve or take longer looks at for further polish:
  * Win detection, while file for tic-tac-toe, is O(N) as it iterates through the statically defined win states for every board. I researched the "magic square" algorithm but the implementation alone (without AI at least) would take a while to create from scratch
  * Similarly, though it could be my game programming naivete, the loops for drawing board states are a place I would spend time examining on whether previous board states could be re-used when drawing
  * As with any library churn and change, there are a few deprecated functions being used that wouldn't normally make it through PR review.
* Definitely overboard on the actual game design aspect. I'm proud of the full pipeline that was completed, but the game loop iterations took enough time that I felt rushed in the timebox

# Other Musings

* I wanted to add a "score" functionality that would incorporate more systemic buildout. For example, your "score" would be the overall time spent to buy 3 stars.
  * With this, I could use Cloudflare/R2 to host a static high score file and people could easily see/share their high scores. I'm unsure whether this is patently unrealistic for 90 mins or not. The important thing is that it would add additional layers to the implementation, for better or worse. Navigating the implementation and API calls to read and update a file for high score listing would be an interesting addition in the scope of time alotted 
* Telemetry for the app would be an awesome addition, but would also require specific tooling. Instrumenting with OpenTelemetry would allow for metrics and traces on overall game performance and player decision-making if this were a "real game". How often upgrades are purchased, etc. Telemetry in anything one builds is just an important consideration
