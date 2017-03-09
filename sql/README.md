## 初期データ生成スクリプト

DBs, Tables 作成
```
mysql -uroot < schema.sql
```

データ投入
```
bundle install
bundle exec ruby seed.rb
```


## 確定データ

webapp/sql を利用する

isuwitter データ投入
```
mysql -uroot -D isuwitter < seed_isuwitter.sql
```

isutomo データ投入
```
mysql -uroot -D isutomo < seed_isutomo.sql
```

## データソース

- name.txt

https://en.wikipedia.org/wiki/Category:Japanese_masculine_given_names
