package com.github.luismoramedina.meshless;

import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;
import org.springframework.context.annotation.Bean;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;
import org.springframework.web.client.RestTemplate;

import java.util.ArrayList;
import java.util.List;

@SpringBootApplication
@RestController
@Slf4j
public class MeshlessBooksApplication {

    @Autowired
    RestTemplate restTemplate;

	@Value("${stars.service.uri}")
	private String url;

	public static void main(String[] args) {
		SpringApplication.run(MeshlessBooksApplication.class, args);
	}

	@Bean
	public RestTemplate restTemplate() {
		return new RestTemplate();
	}

	@RequestMapping
	public List<Book> books() {
		log.info("Before calling " + url);
		Star stars = restTemplate.getForObject(url, Star.class, 1);

		ArrayList<Book> books = new ArrayList<>();
		Book endersGame = new Book();
		endersGame.author = "orson scott card";
		endersGame.title = "Enders game";
		endersGame.year = "1985";
		endersGame.stars = stars.number;
		books.add(endersGame);
		return books;
	}
}
