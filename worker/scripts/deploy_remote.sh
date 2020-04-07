printf "Remote deploy project\n";
eval `ssh-agent -s`;
ssh-add - <<< "${SSH_KEY_REMOTE}";
if [ -z "$ARTIFACT_DIR" ]
then
      scp -rp builds/project-$ID/* $USER@$HOST:$DEPLOY_DIR
else
      scp -rp builds/project-$ID/$ARTIFACT_DIR/* $USER@$HOST:$DEPLOY_DIR
fi
exit 0;