# Tearing eval.Evaluator apart

* Which source file is being evaluated (debugging)
    - May change on closure invoking
* Which node is being evaluated (debugging)
    - Changes now and then, requires a stack
* Maintain variable environment
    - Always changes on closure invoking

* Make up the boundary of Failure
    - Pipeline level
* Make up an executing thread
    - Closure level
