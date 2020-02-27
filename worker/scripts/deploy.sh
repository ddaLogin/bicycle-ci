printf "Deploy project\n";
if [ -z "$ARTIFACT_DIR" ]
then
      cp -a builds/project-$ID/* $DEPLOY_DIR
else
      cp -a builds/project-$ID/$ARTIFACT_DIR/* $DEPLOY_DIR
fi
exit 0;