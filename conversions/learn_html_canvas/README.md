# Converting a JavaScript tutorial to Go

In this repo, I walk you through the steps I took to take Hunor's JavaScript game and convert it to a Web Assembly driven game powered mostly be Go, using [GoDom](https://github.com/ewaldhorn/godom). Cheeky, right?

## Original source
This is loosely based on the excellent video by [Hunor Márton Borbély](https://hunormarton.com/) titled [Learn HTML canvas by coding a JavaScript game | HTML Canvas tutorial for beginners](https://www.youtube.com/watch?v=Ymbv6m3EuNw). In this video, he provides a detailed walkthrough on how to get started building a fun little HTML canvas game. If you are interesting in JavaScript, HTML and CSS, you really should check out his content.

## What we'll be doing

The idea is to take the original JavaScript and place as much of it in Go as possible, with only a thin JavaScript wrapper remaining, because, hey, you can't web without JavaScript!

## How we'll be doing it

I'll be doing it in steps, with each step building on the previous, except for Step 1 of course... You'll need a recent-ish version of Go, I'm rocking 1.27 in November 2026. Other than that, some internet access to download the GoDom package will be most handy, unless you've cloned the repo then you can use relative paths as well, your call.

In order to test our project, we will need some kind of a simple HTTP server. I like [http-server](https://www.npmjs.com/package/http-server) because it's fast to install and simple to use. Pick whatever works for you though.

## Where are the steps

Errr, there are directories for them, I hope?! Try to get to [Step 1](./step_1).

## What's in the directories?

There should be a README.md for each step, as well as some source code and support files. 

## What's this Task business?

Oh yes, the root directory contains a Task file, which is used by [Task](https://taskfile.dev/), a really awesome Go task runner. It's a cross-platform build tool, kind of like Make, but less likely to cut your fingers. You can get started by installing Task and then typing `task` in your terminal in the root directory and it should display your options.
