printf "Go to the project directory\n";
cd builds/project-$ID;
printf "Starting build...\n";
bash <<< "$PLAN"
if [ $? -eq 0 ]; then
    printf "Build was successful\n";
else
    printf "An error occurred while building the project\n";
    exit 1;
fi
exit 0;