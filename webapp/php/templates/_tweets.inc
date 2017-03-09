<?php foreach($tweets as $tweet): ?>
<div class="tweet" data-time="<?= $tweet['time'] ?>">
    <p><a href="/<?= $tweet['user_name'] ?>" class="tweet-user-name"><?= $tweet['user_name'] ?></a></p>
    <p><?= $tweet['html'] ?></p>
    <p class="time"><?= $tweet['time']?></p>
  </div>
<?php endforeach; ?>
