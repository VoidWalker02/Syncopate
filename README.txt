Syncopate is a Go application that watches a directory and mirrors any changes made to it to a second directory.

My personal use case for it is keeping backups of my code in a separate directory when I am programming, ensuring I don't have
just a single copy of the work I've done.

To use it, just type

./Syncopate directory1 directory2

Directory 1 corresponds to the directory you want tracked, directory 2 is where you want it mirrored.
Syncopate does not back up existing files on the mirror directory, but it does mirror any recently created files, I am planning
on supporting that feature on a later version.