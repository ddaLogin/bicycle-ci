printf "Go to the project directory\n";
cd /app;
printf "Starting build...\n";
echo $1 > instruction.sh;
sh instruction.sh;
if [ $? -eq 0 ]; then
    printf "Build was successful\n";
else
    printf "An error occurred while building the project\n";
    exit 1;
fi
exit 0;