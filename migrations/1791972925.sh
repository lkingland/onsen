echo "Set up voxtype dictation the way omarchy install dictation voxtype does"

# voxtype-bin and wtype are declared packages; the stock installer adds the
# model download, the user unit, and makes voxtype the dictation backend.
# Install the packages first, so its gum prompt (shown only when voxtype is
# missing) never comes up in an unattended run.
omarchy-pkg-add wtype voxtype-bin
omarchy-install-dictation-voxtype
