# Creating the containerz RPM and SWIX

Install required packages:

`sudo apt-get install swig rpm`

Now let's create the docker container for containerz

`sh containerize.sh`

Next let's create the RPM and SWIX packages.

`sh rpmize.sh`

