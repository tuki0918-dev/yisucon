<?php include_once("base_top.inc");
if (isset($name) && $name != ''):
   include_once("post.inc"); ?>
   <div class="timeline">
<?php include_once("_tweets.inc"); ?>
   </div>
   <button class="readmore">さらに読み込む</button>
<?php else: ?>
<?php if ($flush != ''): ?>
<p class="flush"><?= $flush ?></p>
<?php endif ?>
   <form class="login" action="/login" method="post">
     <input type="text" name="name">
     <input type="password" name="password">
     <button type="submit">ログイン</button>
   </form>
<?php endif;
include_once("base_bottom.inc") ?>
