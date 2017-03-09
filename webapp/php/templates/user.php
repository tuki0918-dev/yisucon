<?php include_once("base_top.inc");

if (isset($name) && $name != ''):
include_once("post.inc");
endif; ?>

<h3><?= $user ?> さんのツイート</h3>

<?php if($mypage): ?>
<h4>あなたのページです</h4>
<?php elseif($is_friend === true): ?>
<form action="/unfollow" method="post">
    <input type="hidden" name="user" value="<?= $user?>">
    <button type="submit" id="user-unfollow-button">アンフォロー</button>
</form>
<?php elseif (isset($name) && $name != ''): ?>
<form action="/follow" method="post">
    <input type="hidden" name="user" value="<?= $user ?>">
    <button type="submit" id="user-follow-button">フォロー</button>
</form>
<?php endif; ?>

<div class="timeline">
    <?php include_once("_tweets.inc") ?>
</div>
<button class="readmore">さらに読み込む</button>
<?php include_once("base_bottom.inc") ?>

