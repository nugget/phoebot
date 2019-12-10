# Useful commands for merchants and shopkeepers

Register shop containers for sale notifications.  This works best if you've
also set up [the link to your Discord account].

```
/tell Phoebot FORSALE "Meaningful Name" x y z x y z
```

If you tell Phoebot to watch a range of chests/barrels for actiity, you'll get
in-game and Discord messages any time a customer adds or removes items from 
those chests in your shop.

Just figure out the start (x, y, z) coordinates and end (x, y, z) coordinates 
of the containers in your shop.

For example, say you've got a shop where you have a wall of chests to sell all
sorts of wood items.  The upper left corner of those chests is at 100, 66, 200
and the lower right corner of those chests is at 110, 64, 200.  You'd do
something like this:

```
/tell Phoebot FORSALE "My Wood Chests" 100 66 200 110 64 200
```

From then on, Phoebot will let you know when your store has activity


[The link to your Discord account]: discord.md
